import { Link } from "@tanstack/react-router"
import { ThemeTogglerButton } from "@PressScopeBd/ui/components/animate-ui/components/buttons/theme-toggler"
import { SidebarTrigger } from "@PressScopeBd/ui/components/sidebar"

export default function Header() {
  const links = [
    { to: "/", label: "Home" },
    { to: "/settings", label: "Settings" },
    { to: "/user-info", label: "Profile" },
  ] as const

  return (
    <div>
      <div className="flex flex-row items-center justify-between px-2 py-1">
        <div className="flex items-center gap-2">
          <SidebarTrigger />
          <nav className="flex gap-4 text-lg">
            {links.map(({ to, label }) => {
              return (
                <Link key={to} to={to}>
                  {label}
                </Link>
              )
            })}
          </nav>
        </div>
        <div className="flex items-center gap-2">
          <ThemeTogglerButton direction="btt" modes={["light", "dark"]} />
        </div>
      </div>
      <hr />
    </div>
  )
}
