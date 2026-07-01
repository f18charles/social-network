import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Events from "./Events";
import { renderWithProviders } from "../test/render";
import { apiFetch } from "../utils/api";

vi.mock("../utils/api", () => ({
  apiFetch: vi.fn(),
}));

describe("Events", () => {
  beforeEach(() => {
    vi.mocked(apiFetch).mockReset();
    vi.spyOn(window, "alert").mockImplementation(() => {});
  });

  it("loads events only for accepted groups", async () => {
    vi.mocked(apiFetch)
      .mockResolvedValueOnce([
        { id: "group-1", title: "Members", status: "accepted" },
        { id: "group-2", title: "Pending", status: "pending_request" },
      ])
      .mockResolvedValueOnce([
        {
          id: "event-1",
          title: "Planning",
          description: "Scope",
          event_date: "2026-07-12T15:30:00Z",
          going_count: 1,
          not_going_count: 0,
          user_rsvp: "going",
        },
      ]);

    renderWithProviders(<Events />);

    expect(await screen.findByText("Planning")).toBeInTheDocument();
    expect(screen.getByText("Members")).toBeInTheDocument();
    expect(apiFetch).not.toHaveBeenCalledWith("/api/groups/group-2/events");
  });

  it("creates an event and refreshes the event list", async () => {
    vi.mocked(apiFetch)
      .mockResolvedValueOnce([{ id: "group-1", title: "Members", status: "accepted" }])
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce({ id: "event-1" })
      .mockResolvedValueOnce([{ id: "group-1", title: "Members", status: "accepted" }])
      .mockResolvedValueOnce([
        {
          id: "event-1",
          title: "Planning",
          description: "Scope",
          event_date: "2026-07-12T15:30:00Z",
          going_count: 1,
          not_going_count: 0,
          user_rsvp: "going",
        },
      ]);

    renderWithProviders(<Events />);

    fireEvent.click(await screen.findByRole("button", { name: "Create Event" }));
    const textboxes = screen.getAllByRole("textbox");
    fireEvent.change(textboxes[0], { target: { value: "Planning" } });
    fireEvent.change(textboxes[1], { target: { value: "Scope" } });
    fireEvent.change(screen.getByLabelText("Event Date & Time"), {
      target: { value: "2026-07-12T15:30" },
    });
    fireEvent.click(screen.getAllByRole("button", { name: "Create Event" }).at(-1));

    await waitFor(() =>
      expect(apiFetch).toHaveBeenCalledWith("/api/groups/group-1/events", {
        method: "POST",
        body: {
          title: "Planning",
          description: "Scope",
          event_date: new Date("2026-07-12T15:30").toISOString(),
        },
      })
    );
    expect(await screen.findByText("Planning")).toBeInTheDocument();
  });

  it("saves RSVP choices", async () => {
    vi.mocked(apiFetch)
      .mockResolvedValueOnce([{ id: "group-1", title: "Members", status: "accepted" }])
      .mockResolvedValueOnce([
        {
          id: "event-1",
          title: "Planning",
          description: "Scope",
          event_date: "2026-07-12T15:30:00Z",
          going_count: 1,
          not_going_count: 0,
          user_rsvp: "going",
        },
      ])
      .mockResolvedValueOnce({})
      .mockResolvedValueOnce([{ id: "group-1", title: "Members", status: "accepted" }])
      .mockResolvedValueOnce([]);

    renderWithProviders(<Events />);

    fireEvent.click(await screen.findByRole("button", { name: "Not Going" }));

    await waitFor(() =>
      expect(apiFetch).toHaveBeenCalledWith("/api/events/event-1/rsvp", {
        method: "POST",
        body: { status: "not_going" },
      })
    );
  });
});
