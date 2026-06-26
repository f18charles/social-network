import { useEffect, useState } from "react";
import { apiFetch } from "../../utils/api";
import FollowRequestCard from "./FollowRequestCard";
import "../../styles/follow-requests-list.css";

const FollowRequestsList = ({ onRequestCountChange }) => {
  const [requests, setRequests] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchRequests = async () => {
    setLoading(true);
    setError("");
    try {
      const data = await apiFetch("/api/followers/pending");
      setRequests(Array.isArray(data) ? data : []);
      onRequestCountChange?.(data?.length ?? 0);
    } catch (err) {
      setError(err.message || "Failed to load follow requests");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRequests();
  }, []);

  const handleAccept = (requestId) => {
    setRequests((prev) => {
      const updated = prev.filter((r) => r.id !== requestId)
      onRequestCountChange?.(updated.length);
      return updated
    });
  };

  const handleReject = (requestId) => {
    setRequests((prev) => {
      const updated = prev.filter((r) => r.id !== requestId)
      onRequestCountChange?.(updated.length);
      return updated
    });
  };

  if (loading) {
    return (
      <div className="follow-requests-list">
        <div className="follow-requests-list__skeleton" />
        <div className="follow-requests-list__skeleton" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="follow-requests-list__error">{error}</div>
    );
  }

  if (requests.length === 0) {
    return (
      <div className="follow-requests-list__empty">
        No pending follow requests.
      </div>
    );
  }

  return (
    <div className="follow-requests-list">
      {requests.map((request) => (
        <FollowRequestCard
          key={request.id}
          request={request}
          onAccept={handleAccept}
          onReject={handleReject}
        />
      ))}
    </div>
  );
};

export default FollowRequestsList;