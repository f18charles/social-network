import { Outlet } from "react-router";
import "../../styles/layout.css";
import Header from "../Header";
import SidebarLeft from "../SidebarLeft";
import SidebarRight from "../SidebarRight";

const Layout = () => {
  return (
    <div className="main-container">
      <SidebarLeft />
      <Header />
      <div className="main-content">
        <Outlet />
      </div>
      <SidebarRight />
    </div>
  );
};

export default Layout;
