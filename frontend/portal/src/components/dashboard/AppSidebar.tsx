// import { useState } from "react";
import {
  // BarChart2,
  // CreditCard,
  // Users
  LayoutDashboard,
  MessageSquare,
  Settings,
  Workflow,
  CreditCard,
  Tag,
  Edit,
  Zap
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
import { routes } from '@doota/ui-core/routing'
import { useAuth } from "@doota/ui-core/hooks/useAuth";
import { Badge } from "@/components/ui/badge";

export function AppSidebar() {
  const location = usePathname();
  const { getPlanDetails } = useAuth();
  const { planName } = getPlanDetails();
  // const [open, setOpen] = useState(true);

  const isActive = (path: string) => {
    return location.startsWith(path);
  };

  const mainMenuItems = [
    {
      title: "Dashboard",
      path: routes.new.dashboard,
      icon: LayoutDashboard,
      active: isActive(routes.new.dashboard),
    },
    {
      title: "Keywords & Subreddits",
      path: routes.new.keywords,
      icon: Tag,
      active: isActive(routes.new.keywords),
    },
    {
      title: "Lead Feed",
      path: routes.new.leads,
      icon: MessageSquare,
      active: isActive(routes.new.leads),
    },
    {
      title: "Interactions",
      path: routes.new.interactions,
      icon: Zap,
      active: isActive(routes.new.interactions),
    },
  ];

  const workspaceSettingsItems = [
    {
      title: "Edit Product",
      path: routes.new.edit_product,
      icon: Edit,
      active: isActive(routes.new.edit_product),
    },
    {
      title: "Integrations",
      path: routes.new.integrations,
      icon: Settings,
      active: isActive(routes.new.integrations),
    },
    {
      title: "DM Automation",
      path: routes.new.automation,
      icon: Workflow,
      active: isActive(routes.new.automation),
    },
    // {
    //   title: "Reddit Accounts",
    //   path: "/settings/reddit-accounts",
    //   icon: BarChart2,
    //   active: isActive("/settings/reddit-accounts"),
    // },
    // {
    //   title: "Team Members",
    //   path: "/settings/team",
    //   icon: Users,
    //   active: isActive("/settings/team"),
    // },
    {
      title: "Billing Plan",
      path: routes.new.billing,
      icon: CreditCard,
      active: isActive(routes.new.billing),
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
            <div className="text-xs flex items-center text-muted-foreground gap-2.5">
              <p>Current Plan:</p>
              <Badge variant={"default"} className="px-1.5 py-0.5 flex items-center">
                <span className="text-xs">{planName}</span>
              </Badge>
            </div>
          </div>
        </div>
      </SidebarFooter>
    </Sidebar>
  );
}
