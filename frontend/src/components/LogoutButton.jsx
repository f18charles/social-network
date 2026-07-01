import { useAuth } from "../context/auth/useAuth";
import { useNavigate } from "react-router";
import { BiLogOut } from "react-icons/bi";

const LogoutButton = ({ className }) => {
  const { logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await logout();
      navigate("/login");
    } catch (error) {
      console.error("Logout failed:", error);
    }
  };

  return (
    <button
      onClick={handleLogout}
      className={className}
      aria-label="Logout"
      title="Logout"
      style={{
        background: "none",
        border: "none",
        cursor: "pointer",
        display: "flex",
        alignItems: "center",
        padding: "0.5rem",
        color: "inherit",
      }}
    >
      <BiLogOut size={24} />
    </button>
  );
};

export default LogoutButton;
