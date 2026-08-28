import { Navigate, useLocation } from "react-router";
import { useAuth } from "../context/auth/useAuth";

const ProtectedRoute = ({ children }) => {
  const { isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return (
      <div
        style={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          height: "100vh",
          fontSize: "1.2rem",
          color: "#666",
        }}
      >
        Loading...
      </div>
    );
  }

  if (!isAuthenticated) {
    // Redirect to the welcome page, preserving the location they were trying to visit
    return <Navigate to="/welcome" state={{ from: location }} replace />;
  }

  return children;
};

export default ProtectedRoute;
