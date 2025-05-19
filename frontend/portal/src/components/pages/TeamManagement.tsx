
import { useState } from "react";
import { Users, UserPlus, Mail, Trash2, Bell, BellOff } from "lucide-react";
import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
import { DashboardFooter } from "@/components/dashboard/DashboardFooter";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { toast } from "@/components/ui/use-toast";
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

type UserRole = "admin" | "editor" | "viewer";
type UserStatus = "active" | "invited";

interface TeamMember {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  status: UserStatus;
  notifyOnNewLead: boolean;
}

const roleBadgeStyles = {
  admin: "bg-primary hover:bg-primary/80",
  editor: "bg-blue-500 hover:bg-blue-600",
  viewer: "bg-gray-500 hover:bg-gray-600",
};

const roleLabelMap = {
  admin: "Admin",
  editor: "Editor",
  viewer: "Viewer",
};

const roleDescriptionMap = {
  admin: "Full access (invite/remove users, manage Reddit accounts, billing)",
  editor: "Can edit keywords, respond to leads, but can't manage members or billing",
  viewer: "Can view leads and analytics but cannot engage or modify anything",
};

export default function TeamManagement() {
  const [members, setMembers] = useState<TeamMember[]>([
    {
      id: "1",
      name: "Jane Smith",
      email: "jane.smith@example.com",
      role: "admin",
      status: "active",
      notifyOnNewLead: true,
    },
    {
      id: "2",
      name: "John Doe",
      email: "john.doe@example.com",
      role: "editor",
      status: "active",
      notifyOnNewLead: false,
    },
    {
      id: "3",
      name: "Alice Jones",
      email: "alice.jones@example.com",
      role: "viewer",
      status: "invited",
      notifyOnNewLead: true,
    },
  ]);

  const [inviteDialogOpen, setInviteDialogOpen] = useState(false);
  const [newInvite, setNewInvite] = useState({
    email: "",
    role: "editor" as UserRole,
    message: "",
  });

  const handleNotificationToggle = (id: string, enabled: boolean) => {
    setMembers(
      members.map((member) =>
        member.id === id ? { ...member, notifyOnNewLead: enabled } : member
      )
    );

    toast({
      title: `Notifications ${enabled ? "enabled" : "disabled"}`,
      description: `Notifications have been ${
        enabled ? "enabled" : "disabled"
      } for this team member.`,
    });
  };

  const handleRemoveMember = (id: string, name: string) => {
    setMembers(members.filter((member) => member.id !== id));
    
    toast({
      title: "Team member removed",
      description: `${name} has been removed from the workspace.`,
    });
  };

  const handleSendInvite = () => {
    if (!newInvite.email.trim()) {
      toast({
        title: "Email required",
        description: "Please enter an email address for the invitation.",
        variant: "destructive",
      });
      return;
    }

    const newMember: TeamMember = {
      id: `invite-${Date.now()}`,
      name: newInvite.email.split('@')[0],
      email: newInvite.email,
      role: newInvite.role,
      status: "invited",
      notifyOnNewLead: true,
    };

    setMembers([...members, newMember]);
    setInviteDialogOpen(false);
    setNewInvite({
      email: "",
      role: "editor",
      message: "",
    });

    toast({
      title: "Invitation sent",
      description: `An invitation has been sent to ${newInvite.email}.`,
    });
  };

  return (
    <div className="min-h-screen flex flex-col bg-gradient-to-b from-background to-secondary/20">
      <DashboardHeader />
      
      <main className="container mx-auto px-4 py-6 md:px-6">
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-primary to-purple-500 bg-clip-text text-transparent">
                Team Management
              </h1>
              <p className="text-muted-foreground mt-2">
                Invite and manage team members for your workspace
              </p>
            </div>
            <Dialog open={inviteDialogOpen} onOpenChange={setInviteDialogOpen}>
              <DialogTrigger asChild>
                <Button className="gap-1">
                  <UserPlus className="h-4 w-4" />
                  Invite New Member
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-md">
                <DialogHeader>
                  <DialogTitle>Invite Team Member</DialogTitle>
                  <DialogDescription>
                    Send an invitation to collaborate in this workspace
                  </DialogDescription>
                </DialogHeader>
                <div className="space-y-4 pt-4">
                  <div className="space-y-2">
                    <Label htmlFor="email">Email address</Label>
                    <Input
                      id="email"
                      type="email"
                      placeholder="colleague@example.com"
                      value={newInvite.email}
                      onChange={(e) => 
                        setNewInvite({ ...newInvite, email: e.target.value })
                      }
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="role">Role</Label>
                    <Select
                      value={newInvite.role}
                      onValueChange={(value) => 
                        setNewInvite({ ...newInvite, role: value as UserRole })
                      }
                    >
                      <SelectTrigger id="role" className="w-full">
                        <SelectValue placeholder="Select a role" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="admin">
                          <div className="flex flex-col">
                            <span className="font-medium">Admin</span>
                            <span className="text-xs text-muted-foreground">
                              {roleDescriptionMap.admin}
                            </span>
                          </div>
                        </SelectItem>
                        <SelectItem value="editor">
                          <div className="flex flex-col">
                            <span className="font-medium">Editor</span>
                            <span className="text-xs text-muted-foreground">
                              {roleDescriptionMap.editor}
                            </span>
                          </div>
                        </SelectItem>
                        <SelectItem value="viewer">
                          <div className="flex flex-col">
                            <span className="font-medium">Viewer</span>
                            <span className="text-xs text-muted-foreground">
                              {roleDescriptionMap.viewer}
                            </span>
                          </div>
                        </SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="message">Personal message (optional)</Label>
                    <Textarea
                      id="message"
                      placeholder="Hey! I'd like you to join our workspace..."
                      value={newInvite.message}
                      onChange={(e) => 
                        setNewInvite({ ...newInvite, message: e.target.value })
                      }
                    />
                  </div>
                </div>
                <DialogFooter className="mt-4">
                  <Button type="button" variant="outline" onClick={() => setInviteDialogOpen(false)}>
                    Cancel
                  </Button>
                  <Button type="button" onClick={handleSendInvite} className="gap-1">
                    <Mail className="h-4 w-4" />
                    Send Invite
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>

          <div className="rounded-md border shadow-sm bg-card">
            {members.length > 0 ? (
              <Table>
                <TableCaption>Active and invited team members for this workspace</TableCaption>
                <TableHeader>
                  <TableRow>
                    <TableHead>User Name</TableHead>
                    <TableHead>Email</TableHead>
                    <TableHead>Role</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Notify on New Lead</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {members.map((member) => (
                    <TableRow key={member.id}>
                      <TableCell className="font-medium">{member.name}</TableCell>
                      <TableCell>{member.email}</TableCell>
                      <TableCell>
                        <Badge className={roleBadgeStyles[member.role]}>
                          {roleLabelMap[member.role]}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {member.status === "invited" ? (
                          <Badge variant="outline" className="border-amber-500 text-amber-500">
                            Invited – Awaiting Acceptance
                          </Badge>
                        ) : (
                          <Badge variant="outline" className="border-green-500 text-green-500">
                            Active
                          </Badge>
                        )}
                      </TableCell>
                      <TableCell>
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <div className="flex items-center">
                                <Switch
                                  checked={member.notifyOnNewLead}
                                  onCheckedChange={(checked) => handleNotificationToggle(member.id, checked)}
                                  className="mr-2"
                                />
                                {member.notifyOnNewLead ? (
                                  <Bell className="h-4 w-4 text-muted-foreground" />
                                ) : (
                                  <BellOff className="h-4 w-4 text-muted-foreground" />
                                )}
                              </div>
                            </TooltipTrigger>
                            <TooltipContent side="top">
                              <p>Get email alerts when new posts match your workspace keywords</p>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      </TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRemoveMember(member.id, member.name)}
                          className="text-destructive hover:text-destructive/90 hover:bg-destructive/10"
                        >
                          <Trash2 className="h-4 w-4 mr-1" />
                          Remove
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            ) : (
              <div className="flex flex-col items-center justify-center p-10 text-center">
                <Users className="h-10 w-10 text-muted-foreground mb-4" />
                <h3 className="text-lg font-medium mb-2">No team members yet</h3>
                <p className="text-muted-foreground mb-4">
                  Invite others to collaborate on Reddit outreach for this workspace.
                </p>
                <Button onClick={() => setInviteDialogOpen(true)} className="gap-1">
                  <UserPlus className="h-4 w-4" />
                  Invite Your First Team Member
                </Button>
              </div>
            )}
          </div>

          <div className="bg-card border rounded-md p-6 shadow-sm">
            <h2 className="text-xl font-semibold mb-4">Role Permissions</h2>
            <div className="grid gap-4 md:grid-cols-3">
              <div className="p-4 rounded-md border bg-background/50">
                <div className="flex items-center mb-2">
                  <Badge className={roleBadgeStyles.admin}>Admin</Badge>
                </div>
                <p className="text-sm text-muted-foreground">{roleDescriptionMap.admin}</p>
              </div>
              <div className="p-4 rounded-md border bg-background/50">
                <div className="flex items-center mb-2">
                  <Badge className={roleBadgeStyles.editor}>Editor</Badge>
                </div>
                <p className="text-sm text-muted-foreground">{roleDescriptionMap.editor}</p>
              </div>
              <div className="p-4 rounded-md border bg-background/50">
                <div className="flex items-center mb-2">
                  <Badge className={roleBadgeStyles.viewer}>Viewer</Badge>
                </div>
                <p className="text-sm text-muted-foreground">{roleDescriptionMap.viewer}</p>
              </div>
            </div>
          </div>
        </div>
      </main>
      
      <DashboardFooter />
    </div>
  );
}
