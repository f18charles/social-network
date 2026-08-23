import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./styles/index.css";
import App from "./App.jsx";
import { BrowserRouter } from "react-router";
import { AuthProvider } from "./context/auth/AuthContext.jsx";
import { installGlobalErrorLogging } from "./utils/logger.js";
import { startPingService } from "./utils/pingService.js";

installGlobalErrorLogging();

// Keep the Render free-tier backend alive by pinging /health every 10 minutes.
startPingService();

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <AuthProvider>
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </AuthProvider>
  </StrictMode>
);
