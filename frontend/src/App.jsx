import { Route, Routes } from "react-router";
import Layout from "./components/layout/Layout";
import Events from "./pages/Events.jsx";
import Groups from "./pages/Groups.jsx";
import Home from "./pages/Home.jsx";
import Friends from "./pages/Friends.jsx";
import Profile from "./pages/Profile.jsx";
import Messages from "./pages/Messages.jsx";
import Notifications from "./pages/Notifications.jsx";
import PostDetail from "./pages/PostDetail.jsx";
import RegisterForm from "./components/RegisterForm";
import LoginForm from "./components/LoginForm";
import ProtectedRoute from "./components/ProtectedRoute";
import { useAuth } from "./context/auth/useAuth.js";
import { SocketProvider } from "./context/socket"

function App() {
  const { isAuthenticated } = useAuth();
  return (
    <SocketProvider>
      <Routes>
        {/* If not authenticated, display this page without <Layout/>*/}
        {!isAuthenticated && <Route path="/post/:id" element={<PostDetail />} />}
        <Route path="/register" element={<RegisterForm />} />
        <Route path="/login" element={<LoginForm />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Home />} />
          {/* If not authenticated, display this page within <Layout/>*/}
          <Route path="/post/:id" element={<PostDetail />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/user/:userId" element={<Profile />} />
          <Route path="/friends" element={<Friends />} />
          <Route path="/events" element={<Events />} />
          <Route path="/groups" element={<Groups />} />
          <Route path="/messages" element={<Messages />} />
          <Route path="/notifications" element={<Notifications />} />
        </Route>
      </Routes>
    </SocketProvider>
  );
}

export default App;
