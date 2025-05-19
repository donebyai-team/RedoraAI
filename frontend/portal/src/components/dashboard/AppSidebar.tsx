// import { useState } from "react";
import { 
  BarChart2, 
  CreditCard, 
  LayoutDashboard, 
  MessageSquare, 
  // Settings, 
  Tag, 
  Users
} from "lucide-react";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  // SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { usePathname } from "next/navigation";
import Link from "next/link";

export function AppSidebar() {
  const location = usePathname();
  // const [open, setOpen] = useState(true);

  const isActive = (path: string) => {
    return location.startsWith(path);
  };
  
  const mainMenuItems = [
    {
      title: "Dashboard",
      path: "/dashboard",
      icon: LayoutDashboard,
      active: isActive("/dashboard"),
    },
    {
      title: "Keywords & Subreddits",
      path: "/keywords",
      icon: Tag,
      active: isActive("/keywords"),
    },
    {
      title: "Lead Feed",
      path: "/leads",
      icon: MessageSquare,
      active: isActive("/leads"),
    },
  ];
  
  const workspaceSettingsItems = [
    {
      title: "Reddit Accounts",
      path: "/settings/reddit-accounts",
      icon: BarChart2,
      active: isActive("/settings/reddit-accounts"),
    },
    {
      title: "Team Members",
      path: "/settings/team",
      icon: Users,
      active: isActive("/settings/team"),
    },
    {
      title: "Billing Plan",
      path: "/settings/billing",
      icon: CreditCard,
      active: isActive("/settings/billing"),
    },
  ];

  return (
    <Sidebar>
      <SidebarHeader className="pb-0">
        <div className="flex items-center justify-between p-2">
          <Link href="/dashboard" className="flex items-center gap-2 px-2">
            <div className="bg-gradient-to-r from-primary to-purple-500 text-white p-1.5 rounded-md">
              <MessageSquare className="h-4 w-4" />
            </div>
            <span className="font-bold text-xl">Redora</span>
          </Link>
          <SidebarTrigger />
        </div>
      </SidebarHeader>
      
      <SidebarContent className="flex-grow">
        <SidebarGroup>
          <SidebarGroupLabel>Main</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {mainMenuItems.map((item) => (
                <SidebarMenuItem key={item.path}>
                  <SidebarMenuButton asChild isActive={item.active}>
                    <Link href={item.path} className="flex items-center">
                      <item.icon className="h-4 w-4 mr-2" />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        
        <SidebarGroup>
          <SidebarGroupLabel>Workspace Settings</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {workspaceSettingsItems.map((item) => (
                <SidebarMenuItem key={item.path}>
                  <SidebarMenuButton asChild isActive={item.active}>
                    <Link href={item.path} className="flex items-center">
                      <item.icon className="h-4 w-4 mr-2" />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      
      <SidebarFooter className="mt-auto">
        <div className="px-3 py-2 border-t border-sidebar-border">
          <div className="flex justify-between items-center">
            <div className="text-xs text-muted-foreground">
              <p>Workspace: Personal</p>
            </div>
          </div>
        </div>
      </SidebarFooter>
    </Sidebar>
  );
}
