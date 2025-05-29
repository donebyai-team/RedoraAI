
import { useState } from "react";
import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
import { DashboardFooter } from "@/components/dashboard/DashboardFooter";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { toast } from "@/components/ui/use-toast";
import { Edit, MoreHorizontal, Plus, Trash2, UserPlus } from "lucide-react";
import { format } from "date-fns";
import Link from "next/link";

interface Workspace {
  id: string;
  name: string;
  createdAt: Date;
  teamSize: number;
  plan: "starter" | "pro" | "enterprise";
}

export default function WorkspacesManagement() {
  const [workspaces, setWorkspaces] = useState<Workspace[]>([
    {
      id: "w1",
      name: "Personal Workspace",
      createdAt: new Date(2023, 5, 15),
      teamSize: 1,
      plan: "starter"
    },
    {
      id: "w2",
      name: "Agency Clients",
      createdAt: new Date(2023, 8, 3),
      teamSize: 5,
      plan: "pro"
    },
    {
      id: "w3",
      name: "Marketing Team",
      createdAt: new Date(2024, 1, 22),
      teamSize: 3,
      plan: "pro"
    },
  ]);

  const [newWorkspaceName, setNewWorkspaceName] = useState("");

  const handleCreateWorkspace = () => {
    if (newWorkspaceName.trim() === "") {
      toast({
        title: "Please enter a workspace name",
        variant: "destructive",
      });
      return;
    }

    const newWorkspace: Workspace = {
      id: `w${Date.now()}`,
      name: newWorkspaceName,
      createdAt: new Date(),
      teamSize: 1,
      plan: "starter",
    };

    setWorkspaces([...workspaces, newWorkspace]);
    setNewWorkspaceName("");

    toast({
      title: "Workspace created",
      description: `${newWorkspaceName} has been created successfully.`,
    });
  };

  const getPlanBadge = (plan: Workspace["plan"]) => {
    switch (plan) {
      case "starter":
        return <Badge variant="secondary">Starter</Badge>;
      case "pro":
        return <Badge variant="outline" className="border-primary/50 text-primary">Pro</Badge>;
      case "enterprise":
        return <Badge className="bg-gradient-to-r from-primary to-purple-500">Enterprise</Badge>;
    }
  };

  const handleDelete = (id: string) => {
    setWorkspaces(workspaces.filter(w => w.id !== id));
    toast({
      title: "Workspace deleted",
      description: "The workspace has been deleted.",
    });
  };

  return (
    <div className="min-h-screen flex flex-col bg-gradient-to-b from-background to-secondary/20">
      <DashboardHeader />

      <div className="flex-1 overflow-auto">
        <main className="container mx-auto px-4 py-6 md:px-6">
          <div className="space-y-2 mb-6">
            <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-primary to-purple-500 bg-clip-text text-transparent">
              Workspace Management
            </h1>
            <p className="text-muted-foreground">
              Manage your workspaces for different clients, projects, or teams.
            </p>
          </div>

          {/* Create new workspace */}
          <Card className="border-primary/10 shadow-md mb-6">
            <CardHeader>
              <CardTitle>Create New Workspace</CardTitle>
              <CardDescription>
                Each workspace has its own keywords, subreddits, and Reddit accounts.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex flex-col sm:flex-row gap-3">
                <Input
                  placeholder="Enter workspace name"
                  value={newWorkspaceName}
                  onChange={(e) => setNewWorkspaceName(e.target.value)}
                  className="flex-1"
                />
                <Button onClick={handleCreateWorkspace} className="gap-1">
                  <Plus className="h-4 w-4" />
                  Create Workspace
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Workspaces list */}
          <div className="space-y-4">
            <h2 className="text-xl font-semibold">Your Workspaces</h2>
            {workspaces.length === 0 ? (
              <Card className="border-dashed">
                <CardContent className="flex flex-col items-center justify-center p-6">
                  <p className="text-muted-foreground mb-4">You don't have any workspaces yet.</p>
                  <Button onClick={handleCreateWorkspace} className="gap-1">
                    <Plus className="h-4 w-4" />
                    Create Your First Workspace
                  </Button>
                </CardContent>
              </Card>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {workspaces.map((workspace) => (
                  <Card key={workspace.id} className="border-primary/10 shadow-sm hover:shadow-md transition-shadow">
                    <CardHeader className="pb-2">
                      <div className="flex justify-between items-start">
                        <div className="space-y-1">
                          <CardTitle className="flex items-center gap-2">
                            {workspace.name}
                          </CardTitle>
                          <CardDescription>
                            Created {format(workspace.createdAt, "MMM d, yyyy")}
                          </CardDescription>
                        </div>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="h-8 w-8">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem className="cursor-pointer">
                              <Edit className="h-4 w-4 mr-2" />
                              Edit
                            </DropdownMenuItem>
                            <DropdownMenuItem className="cursor-pointer">
                              <UserPlus className="h-4 w-4 mr-2" />
                              Invite Team Member
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              className="cursor-pointer text-destructive"
                              onClick={() => handleDelete(workspace.id)}
                            >
                              <Trash2 className="h-4 w-4 mr-2" />
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </div>
                    </CardHeader>
                    <CardContent>
                      <div className="flex justify-between pt-2">
                        <div className="space-y-2">
                          <div className="flex items-center gap-1">
                            <span className="text-sm font-medium">Team size:</span>
                            <span className="text-sm text-muted-foreground">{workspace.teamSize}</span>
                          </div>
                          <div>
                            {getPlanBadge(workspace.plan)}
                          </div>
                        </div>
                        <Button size="sm" asChild variant="secondary">
                          <Link href={`/workspaces/${workspace.id}`}>
                            Manage
                          </Link>
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </div>
        </main>
      </div>

      <DashboardFooter />
    </div>
  );
}
