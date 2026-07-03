import { expect, request as playwrightRequest, test } from "@playwright/test";

const backendURL = `http://127.0.0.1:${process.env.E2E_BACKEND_PORT || "18080"}`;
const password = "Password123!";

const unique = (label) => `${label}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;

async function apiContext() {
  return playwrightRequest.newContext({ baseURL: backendURL });
}

async function unwrap(response) {
  const body = await response.json();
  return body?.data ?? body;
}

async function registerUser(api, overrides = {}) {
  const suffix = unique("user");
  const profile = {
    email: `${suffix}@example.test`,
    password,
    first_name: overrides.firstName || "E2E",
    last_name: overrides.lastName || "User",
    date_of_birth: "1995-05-15",
    nickname: overrides.nickname || suffix,
    about_me: overrides.aboutMe || "Created by Playwright e2e.",
    is_public: overrides.isPublic ? "true" : "false",
  };
  const response = await api.post("/api/users/register", { multipart: profile });
  expect(response.status()).toBe(201);
  const user = await unwrap(response);
  return { ...user, email: profile.email, password, nickname: profile.nickname };
}

async function loginApi(api, user) {
  const response = await api.post("/api/users/login", {
    data: { email: user.email, password: user.password },
  });
  expect(response.status()).toBe(200);
}

async function loginBrowser(page, user) {
  await page.goto("/login");
  await page.getByLabel("Email Address").fill(user.email);
  await page.getByLabel("Password").fill(user.password);
  await page.getByRole("button", { name: "Login" }).click();
  await expect(page).toHaveURL(/\/$/);
  await expect(page.getByPlaceholder("What\x27s on your mind?")).toBeVisible();
}

async function createPost(api, { content, privacy = "public", audienceIDs = [], groupID = null }) {
  const multipart = { content, privacy };
  if (audienceIDs.length > 0) multipart.audience_ids = JSON.stringify(audienceIDs);
  if (groupID) multipart.group_id = groupID;
  const response = await api.post("/api/posts", { multipart });
  expect(response.status()).toBe(201);
  return unwrap(response);
}

async function follow(api, targetUserID) {
  const response = await api.post("/api/followers/follow", {
    data: { following_id: targetUserID },
  });
  expect([200, 202]).toContain(response.status());
}

async function acceptFollow(api, followerID) {
  const response = await api.post("/api/followers/accept", {
    data: { follower_id: followerID },
  });
  expect(response.status()).toBe(200);
}

async function createGroup(api, title) {
  const response = await api.post("/api/groups", {
    data: { title, description: "E2E group description" },
  });
  expect(response.status()).toBe(201);
  return unwrap(response);
}

test("registration, login, session, and logout work through the browser", async ({ page }) => {
  const suffix = unique("browser-auth");
  const email = `${suffix}@example.test`;

  await page.goto("/register");
  await page.getByLabel("Email Address *").fill(email);
  await page.getByLabel("Password *").fill(password);
  await page.getByLabel("First Name *").fill("Browser");
  await page.getByLabel("Last Name *").fill("Auth");
  await page.getByLabel("Date of Birth *").fill("1994-04-12");
  await page.getByLabel("Nickname").fill(suffix);
  await page.getByLabel("Public profile").check();
  await page.getByRole("button", { name: "Register" }).click();

  // On registration success, they are automatically logged in and redirected to home feed
  await expect(page.getByPlaceholder("What\x27s on your mind?")).toBeVisible();
  await expect(page).toHaveURL(/\/$/);

  // Click the avatar photo to open the dropdown menu
  await page.getByAltText("Your profile").click();

  // Click the Sign Out button
  await page.getByRole("button", { name: "Sign Out" }).click();

  // Verify they are redirected to /login
  await expect(page).toHaveURL(/\/login$/);

  // Try going back to / and check redirection to /login
  await page.goto("/");
  await expect(page).toHaveURL(/\/login$/);
});

test("private profiles and posts unlock only after follow acceptance", async ({ page }) => {
  const viewerApi = await apiContext();
  const ownerApi = await apiContext();
  const viewer = await registerUser(viewerApi, { firstName: "Viewer", isPublic: true });
  const owner = await registerUser(ownerApi, { firstName: "Private", lastName: "Owner", isPublic: false });
  await loginApi(viewerApi, viewer);
  await loginApi(ownerApi, owner);
  await createPost(ownerApi, { content: "Private owner profile post", privacy: "public" });

  await loginBrowser(page, viewer);
  await page.goto(`/user/${owner.id}`);
  await expect(page.getByText("Private Owner")).toBeVisible();
  await expect(page.getByText("Follow to see their posts")).toBeVisible();
  await expect(page.getByText("Private owner profile post")).toHaveCount(0);

  await page.getByRole("button", { name: "Follow", exact: true }).click();
  await expect(page.getByRole("button", { name: "Requested" })).toBeVisible();
  await acceptFollow(ownerApi, viewer.id);

  await page.reload();
  await expect(page.getByText("Private owner profile post")).toBeVisible();
});

test("group join approval, event creation, RSVP, and notifications follow business rules", async ({ page }) => {
  const creatorApi = await apiContext();
  const memberApi = await apiContext();
  const creator = await registerUser(creatorApi, { firstName: "Group", lastName: "Creator", isPublic: true });
  const member = await registerUser(memberApi, { firstName: "Group", lastName: "Member", isPublic: true });
  await loginApi(creatorApi, creator);
  await loginApi(memberApi, member);
  const group = await createGroup(creatorApi, `E2E Group ${unique("group")}`);

  const joinResponse = await memberApi.post(`/api/groups/${group.id}/join`);
  expect(joinResponse.status()).toBe(200);

  await loginBrowser(page, creator);
  await page.goto("/groups");
  await expect(page.getByText(group.title)).toBeVisible();
  await page.getByRole("button", { name: "Review Join Requests" }).click();
  await expect(page.getByText(member.email)).toBeVisible();
  await page.getByRole("button", { name: "Accept" }).click();
  await expect(page.getByText(member.email)).toHaveCount(0);

  await page.goto("/events");
  await page.getByRole("button", { name: "Create Event" }).click();
  await page.getByLabel("Event Title").fill("E2E Planning Event");
  await page.getByLabel("Description").fill("Validate event flow");
  await page.getByLabel("Event Date & Time").fill("2026-07-12T15:30");
  await page.getByRole("button", { name: "Create Event" }).last().click();
  await expect(page.getByRole("heading", { name: "E2E Planning Event" })).toBeVisible();

  await loginBrowser(page, member);
  await page.goto("/events");
  await expect(page.getByRole("heading", { name: "E2E Planning Event" })).toBeVisible();
  await page.getByRole("button", { name: "Not Going" }).click();
  await expect(page.getByText(/1 Not Going/)).toBeVisible();

  await page.goto("/notifications");
  await expect(page.getByText(/A new event 'E2E Planning Event' was created in your group\./)).toBeVisible();
  await page.getByRole("button", { name: "Mark Read" }).click();
  await expect(page.getByRole("button", { name: "Mark Read" })).toHaveCount(0);
});

test("chat appears only after an allowed follower relationship creates a conversation", async ({ page }) => {
  const senderApi = await apiContext();
  const recipientApi = await apiContext();
  const sender = await registerUser(senderApi, { firstName: "Chat", lastName: "Sender", isPublic: true });
  const recipient = await registerUser(recipientApi, { firstName: "Chat", lastName: "Recipient", isPublic: true });
  await loginApi(senderApi, sender);
  await loginApi(recipientApi, recipient);

  const blocked = await senderApi.post("/api/messages", {
    data: { recipient_id: recipient.id, content: "blocked" },
  });
  expect(blocked.status()).toBe(400);

  await follow(senderApi, recipient.id);
  const sent = await senderApi.post("/api/messages", {
    data: { recipient_id: recipient.id, content: "Allowed DM from sender" },
  });
  expect(sent.status()).toBe(201);

  await loginBrowser(page, recipient);
  await page.goto("/messages");
  await expect(page.getByText("Chat Sender")).toBeVisible();
  await page.locator(".messages-conversation-item-preview").first().click();
  const messageList = page.locator(".messages-list");
  await expect(messageList.getByText("Allowed DM from sender")).toBeVisible();
  await page.getByPlaceholder("Type your message...").fill("Reply from recipient");
  await page.getByRole("button", { name: "Send" }).click();
  await expect(messageList.getByText("Reply from recipient")).toBeVisible();
});
