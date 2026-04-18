import { Outlet } from "react-router";
import NavBar from "@/components/NavBar";

export default function Layout() {
  return (
    <div className="min-h-screen bg-background">
      <NavBar />
      <Outlet />
    </div>
  );
}
