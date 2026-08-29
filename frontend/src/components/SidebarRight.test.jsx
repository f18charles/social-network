import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SidebarRight from "./SidebarRight";
import { renderWithProviders } from "../test/render";
import { apiFetch } from "../utils/api";

vi.mock("../utils/api", () => ({
  apiFetch: vi.fn(),
}));

describe("SidebarRight", () => {
  beforeEach(() => {
    vi.mocked(apiFetch).mockReset();
  });

  it("shows loading state initially and renders upcoming future events only", async () => {
    const futureDate1 = new Date(Date.now() + 86400000).toISOString(); // tomorrow
    const futureDate2 = new Date(Date.now() + 172800000).toISOString(); // 2 days later
    const pastDate = new Date(Date.now() - 86400000).toISOString(); // yesterday

    vi.mocked(apiFetch)
      .mockResolvedValueOnce([
        { id: "group-1", title: "React Developers", status: "accepted" },
        { id: "group-2", title: "Python Group", status: "pending_request" },
      ])
      .mockResolvedValueOnce([
        {
          id: "event-past",
          title: "Past Hackathon",
          description: "Old event",
          event_date: pastDate,
        },
        {
          id: "event-future-2",
          title: "Advanced Workshop",
          description: "Future event 2",
          event_date: futureDate2,
        },
        {
          id: "event-future-1",
          title: "Kickoff Meeting",
          description: "Future event 1",
          event_date: futureDate1,
        },
      ]);

    renderWithProviders(<SidebarRight />);

    expect(screen.getByText("Loading events...")).toBeInTheDocument();

    // Verify only future events are displayed in chronological order
    expect(await screen.findByText("Kickoff Meeting")).toBeInTheDocument();
    expect(screen.getByText("Advanced Workshop")).toBeInTheDocument();
    expect(screen.queryByText("Past Hackathon")).not.toBeInTheDocument();

    // Verify unaccepted group was ignored
    expect(apiFetch).not.toHaveBeenCalledWith("/api/groups/group-2/events");
  });

  it("renders empty state when there are no upcoming events", async () => {
    vi.mocked(apiFetch)
      .mockResolvedValueOnce([
        { id: "group-1", title: "Designers", status: "accepted" },
      ])
      .mockResolvedValueOnce([]);

    renderWithProviders(<SidebarRight />);

    expect(
      await screen.findByText("No upcoming events yet.")
    ).toBeInTheDocument();
  });

  it("handles fetch errors gracefully", async () => {
    vi.mocked(apiFetch).mockRejectedValueOnce(new Error("Network connection error"));

    renderWithProviders(<SidebarRight />);

    expect(
      await screen.findByText("Network connection error")
    ).toBeInTheDocument();
  });

  it("has a link to /events in the header", async () => {
    vi.mocked(apiFetch).mockResolvedValueOnce([]);

    renderWithProviders(<SidebarRight />);

    const viewAllLink = await screen.findByRole("link", { name: "View all" });
    expect(viewAllLink).toHaveAttribute("href", "/events");
  });
});
