
import { useState } from "react";
import {
  ChevronDown,
  // Plus, 
  // Settings 
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  // DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
// import Link from "next/link";
import { useAuth } from "@doota/ui-core/hooks/useAuth";
import { Project } from "@doota/pb/doota/core/v1/core_pb";
import { isPlatformAdmin } from "@doota/ui-core/helper/role";
import { useOrganization } from "@doota/ui-core/hooks/useOrganization";

interface Workspace {
  id: string;
  name: string;
}

export function WorkspaceSwitcher() {

  const { user } = useAuth()
  const [currentOrg, setCurrentOrganization] = useOrganization();

  const canChangeOrg = user && isPlatformAdmin(user) && user.organizations.length > 1;
  const workspaces: Project[] = user?.projects ?? [];
  const [currentWorkspace, setCurrentWorkspace] = useState<Workspace>(workspaces[0]);

  return (<>
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="sm" className="h-9 gap-1 px-2">
          <span className="text-sm font-normal text-muted-foreground mr-1">
            Workspace:
          </span>
          <span className="text-sm font-medium max-w-[150px] truncate">
            {currentWorkspace?.name}
          </span>
          <ChevronDown className="h-4 w-4 text-muted-foreground" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-[220px]">
        {workspaces.map((workspace) => (
          <DropdownMenuItem
            key={workspace.id}
            onClick={() => setCurrentWorkspace(workspace)}
            className="cursor-pointer flex items-center justify-between"
          >
            <span className="truncate">{workspace?.name}</span>
            {currentWorkspace.id === workspace.id && (
              <span className="w-2 h-2 rounded-full bg-primary ml-2"></span>
            )}
          </DropdownMenuItem>
        ))}
        {/* <DropdownMenuSeparator />
        <DropdownMenuItem asChild>
          <Link href="/workspaces/new" className="cursor-pointer flex items-center gap-2">
            <Plus className="h-4 w-4" />
            <span>Create New Workspace</span>
          </Link>
        </DropdownMenuItem>
        <DropdownMenuItem asChild>
          <Link href="/workspaces" className="cursor-pointer flex items-center gap-2">
            <Settings className="h-4 w-4" />
            <span>Manage All Workspaces</span>
          </Link>
        </DropdownMenuItem> */}
      </DropdownMenuContent>
    </DropdownMenu>

    {user && isPlatformAdmin(user) && (
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="sm" className="h-9 gap-1 px-2 ml-2">
            <span className="text-sm font-medium max-w-[150px] truncate">
              {currentOrg?.name}
            </span>
            {canChangeOrg && (
              <ChevronDown className="h-4 w-4 text-muted-foreground" />
            )}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="max-h-[300px] overflow-y-auto">
          {[...user.organizations]
            .sort((a, b) => a.name.localeCompare(b.name))
            .map((workspace) => (
              <DropdownMenuItem
                key={workspace.id}
                onClick={() => {
                  setCurrentOrganization(workspace).then(() => {
                    window.location.reload();
                  });
                }}
                className="cursor-pointer flex items-center justify-between"
              >
                <span className="truncate">{workspace?.name}</span>
                {currentOrg?.id === workspace.id && (
                  <span className="w-2 h-2 rounded-full bg-primary ml-2"></span>
                )}
              </DropdownMenuItem>
            ))}
        </DropdownMenuContent>

      </DropdownMenu>
    )}
  </>);
}
