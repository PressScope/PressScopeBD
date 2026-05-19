import { Link } from "@tanstack/react-router";
import { ThemeTogglerButton } from "@PressScopeBd/ui/components/animate-ui/components/buttons/theme-toggler";
import { SidebarTrigger } from "@PressScopeBd/ui/components/sidebar";

export default function Header() {
  const isLoginPage = window.location.pathname === "/login";
  return (
    <div>
      <div className="flex flex-row items-center justify-between px-2 py-1">
        <div className="flex items-center gap-2">
          {!isLoginPage && <SidebarTrigger />}
        </div>
        <div className="flex items-center gap-2">
          <ThemeTogglerButton direction="btt" modes={["light", "dark"]} />
        </div>
      </div>
      <hr />
    </div>
  );
}
