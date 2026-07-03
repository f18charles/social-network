import { useState } from "react";
import { Outlet } from "react-router";
import "../../styles/layout.css";
import Header from "../Header";
import SidebarLeft from "../SidebarLeft";
import SidebarRight from "../SidebarRight";

const Layout = () => {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <div
      className={`main-container ${isCollapsed ? "sidebar-collapsed" : ""} ${
        mobileOpen ? "mobile-menu-open" : ""
      }`}
    >
      <SidebarLeft
        isCollapsed={isCollapsed}
        onToggleCollapse={() => setIsCollapsed(!isCollapsed)}
      />
      <Header onToggleMobileMenu={() => setMobileOpen(!mobileOpen)} />
      <div className="main-content">
        <Outlet />
      </div>
      <SidebarRight />
    </div>
  );
};

export default Layout;
