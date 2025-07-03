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
  Zap,
  X,
  PanelLeft,
  HelpCircle,
  TrendingUp
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
  useSidebar,
} from "@/components/ui/sidebar";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { routes } from '@doota/ui-core/routing'
import { useAuth } from "@doota/ui-core/hooks/useAuth";
import { Badge } from "@/components/ui/badge";
import { Image } from "@doota/ui-core/atoms/Image";
import { getFreePlanDateStatus } from "@/utils/format";
import { Button } from "../ui/button";
import { SubscriptionPlanID } from "@doota/pb/doota/core/v1/core_pb";

export function AppSidebar() {

  const { isMobile, toggleSidebar, openMobile } = useSidebar();

  const location = usePathname();
  const { getPlanDetails, currentOrganization } = useAuth();
  const { planId } = getPlanDetails();
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
      title: "Insights",
      path: routes.new.insights,
      icon: TrendingUp,
      active: isActive(routes.new.insights),
    },
    {
      title: "Keywords",
      path: routes.new.keywords,
      icon: Tag,
      active: isActive(routes.new.keywords),
    },
    {
      title: "Tracked Conversations",
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
      icon: Workflow,
      active: isActive(routes.new.integrations),
    },
    {
      title: "Automation Settings",
      path: routes.new.automation,
      icon: Settings,
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
    {
      title: "Help Center",
      path: "https://redoraai.featurebase.app/help",
      icon: HelpCircle, // Or another appropriate icon
      external: true,
    }
  ];

  function getPlanSuffix(planId: SubscriptionPlanID): string {
    const key = SubscriptionPlanID[planId]; // e.g. "SUBSCRIPTION_PLAN_FREE"
    return key.replace('SUBSCRIPTION_PLAN_', ''); // → "FREE"
  }

  return (
    <Sidebar>
      <SidebarHeader className="pb-0">
        <div className="flex items-center justify-between p-2">
          <Link href="/dashboard" className="flex items-center gap-2 px-2">
            <div className="text-white mr-1">
              <Image width={35} height={35} alt='doota logo' priority imageKey='logo_circle' />
            </div>
            <span className="font-bold text-xl">Redora</span>
          </Link>
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={toggleSidebar}
          >
            {isMobile && openMobile ? (
              <X className="h-4 w-4" />
            ) : (
              <PanelLeft className="h-4 w-4" />
            )}
          </Button>
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
                    {item.external ? (
                      <a
                        href={item.path}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center"
                      >
                        <item.icon className="h-4 w-4 mr-2" />
                        <span>{item.title}</span>
                      </a>
                    ) : (
                      <Link href={item.path} className="flex items-center">
                        <item.icon className="h-4 w-4 mr-2" />
                        <span>{item.title}</span>
                      </Link>
                    )}
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
              <Badge variant="default" className="px-2 py-1 flex items-center rounded-md">
                <span className="text-xs">
                  {getPlanSuffix(planId)}
                  {planId === SubscriptionPlanID.SUBSCRIPTION_PLAN_FREE &&
                    ` ${getFreePlanDateStatus(currentOrganization?.featureFlags?.subscription?.expiresAt)}`}
                </span>
              </Badge>
            </div>
          </div>
        </div>
      </SidebarFooter>

    </Sidebar>
  );
}
