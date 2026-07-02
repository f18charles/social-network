import { useNavigate } from "react-router";
import ProfileUpdateForm from "../components/ProfileUpdateForm";

const EditProfile = () => {
  const navigate = useNavigate();
  return (
    <div className="edit-profile-page" style={{ padding: "2rem", maxWidth: "600px", margin: "0 auto" }}>
      <ProfileUpdateForm onClose={() => navigate("/profile")} />
    </div>
  );
};

export default EditProfile;
