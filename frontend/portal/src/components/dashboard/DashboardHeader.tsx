
import React from "react";
import { Button } from "@/components/ui/button";
import {
  // Bell,
  Settings,
  LogOut,
  User,
} from "lucide-react";
import { WorkspaceSwitcher } from "@/components/workspace/WorkspaceSwitcher";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import Link from "next/link";
import { useAuth } from "@doota/ui-core/hooks/useAuth";
import { routes } from "@doota/ui-core/routing";

export function DashboardHeader() {

  const { logout } = useAuth();

  const handleLogout = () => {
    logout()
  };

  return (
    <header className="border-b border-primary/10 bg-background/95 py-3 px-4 md:px-6">
      <div className="container mx-auto">
        <div className="flex items-center justify-between">
          {/* Left section - Workspace Switcher */}
          <div className="flex items-center space-x-4">
            <WorkspaceSwitcher />
          </div>

          {/* Right section - Notifications and Profile */}
          <div className="flex items-center gap-1 md:gap-2">
            {/* <Button variant="ghost" size="icon" className="relative">
              <Bell className="h-4 w-4" />
              <span className="absolute top-1 right-1 w-2 h-2 rounded-full bg-primary" />
            </Button> */}

            {/* Profile Dropdown */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon">
                  <div className="h-7 w-7 rounded-full bg-secondary flex items-center justify-center">
                    <User className="h-4 w-4" />
                  </div>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuLabel>My Account</DropdownMenuLabel>
                <DropdownMenuSeparator />
                {/* <DropdownMenuItem asChild>
                  <Link href="/profile" className="cursor-pointer">
                    <User className="h-4 w-4 mr-2" />
                    Profile
                  </Link>
                </DropdownMenuItem> */}
                <DropdownMenuItem asChild>
                  <Link href={routes.new.integrations} className="cursor-pointer">
                    <Settings className="h-4 w-4 mr-2" />
                    Global Settings
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={handleLogout}>
                  <LogOut className="h-4 w-4 mr-2" />
                  Logout
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>
    </header>
  );
}
