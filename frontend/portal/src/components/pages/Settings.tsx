
import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Slider } from "@/components/ui/slider";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Settings, Sliders, Bell, User } from "lucide-react";
import { DashboardHeader } from "@/components/dashboard/DashboardHeader";
import { toast } from "@/components/ui/use-toast";

export default function SettingsPage() {
  const [relevancyScore, setRelevancyScore] = useState(75);
  const [autoSave, setAutoSave] = useState(true);
  const [autoComment, setAutoComment] = useState(false);
  const [autoDM, setAutoDM] = useState(false);
  const [emailNotifications, setEmailNotifications] = useState(true);
  const [pushNotifications, setPushNotifications] = useState(true);
  
  const handleSaveGeneral = () => {
    toast({
      title: "Settings saved",
      description: "Your general settings have been updated.",
    });
  };
  
  const handleSaveAutomation = () => {
    toast({
      title: "Automation updated",
      description: "Your automation settings have been saved.",
    });
  };
  
  const handleSaveNotifications = () => {
    toast({
      title: "Notifications updated",
      description: "Your notification preferences have been saved.",
    });
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-background to-secondary/20">
      <DashboardHeader />
      
      <main className="container mx-auto px-4 py-6 md:px-6">
        <div className="space-y-6">
          <div>
            <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-primary to-purple-500 bg-clip-text text-transparent">Settings</h1>
            <p className="text-muted-foreground mt-2">
              Customize how Redora AI works for you.
            </p>
          </div>
          
          <Tabs defaultValue="general" className="space-y-4">
            <TabsList className="bg-secondary/50">
              <TabsTrigger 
                className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary" 
                value="general"
              >
                General
              </TabsTrigger>
              <TabsTrigger 
                className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary" 
                value="automation"
              >
                Automation
              </TabsTrigger>
              <TabsTrigger 
                className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary" 
                value="notifications"
              >
                Notifications
              </TabsTrigger>
              <TabsTrigger 
                className="data-[state=active]:bg-primary/10 data-[state=active]:text-primary" 
                value="account"
              >
                Account
              </TabsTrigger>
            </TabsList>
            
            <TabsContent value="general" className="space-y-4">
              <Card className="border-primary/10 shadow-md">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Settings className="h-5 w-5" />
                    General Settings
                  </CardTitle>
                  <CardDescription>Configure your general preferences.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="username">Username</Label>
                      <Input id="username" defaultValue="redora_user" />
                    </div>
                    
                    <div className="space-y-2">
                      <Label htmlFor="email">Email Address</Label>
                      <Input id="email" type="email" defaultValue="user@example.com" />
                    </div>
                    
                    <div className="space-y-2">
                      <Label htmlFor="timezone">Timezone</Label>
                      <select 
                        id="timezone" 
                        className="w-full flex h-10 rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                      >
                        <option value="UTC-8">Pacific Time (UTC-8)</option>
                        <option value="UTC-5">Eastern Time (UTC-5)</option>
                        <option value="UTC">UTC</option>
                        <option value="UTC+1">Central European Time (UTC+1)</option>
                        <option value="UTC+8">China Standard Time (UTC+8)</option>
                      </select>
                    </div>
                    
                    <Button onClick={handleSaveGeneral} className="bg-gradient-to-r from-primary to-purple-500 hover:from-primary/90 hover:to-purple-500/90">
                      Save Changes
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
            
            <TabsContent value="automation" className="space-y-4">
              <Card className="border-primary/10 shadow-md">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Sliders className="h-5 w-5" />
                    Automation Settings
                  </CardTitle>
                  <CardDescription>Configure your automation preferences.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div className="space-y-6">
                    <div className="space-y-2">
                      <div className="flex items-center justify-between">
                        <Label>Minimum Relevancy Score: {relevancyScore}%</Label>
                        <span className={`text-sm font-semibold ${
                          relevancyScore >= 90 ? "text-green-500" : 
                          relevancyScore >= 70 ? "text-amber-500" : 
                          "text-red-500"
                        }`}>
                          {relevancyScore >= 90 ? "Excellent" : 
                           relevancyScore >= 70 ? "Good" : 
                           relevancyScore >= 50 ? "Moderate" : "Poor"}
                        </span>
                      </div>
                      <Slider
                        defaultValue={[relevancyScore]} 
                        max={100} 
                        step={1}
                        onValueChange={(value) => setRelevancyScore(value[0])}
                        className="bg-gradient-to-r from-red-500 via-amber-500 to-green-500 h-2 rounded-full"
                      />
                      <p className="text-xs text-muted-foreground">Posts with scores below this threshold won't trigger automation.</p>
                    </div>
                    
                    <div className="space-y-4">
                      <div className="flex items-center space-x-2">
                        <Switch 
                          id="auto-save" 
                          checked={autoSave}
                          onCheckedChange={setAutoSave}
                        />
                        <Label htmlFor="auto-save">Auto-save high-scoring posts</Label>
                      </div>
                      
                      <div className="flex items-center space-x-2">
                        <Switch 
                          id="auto-comment" 
                          checked={autoComment}
                          onCheckedChange={setAutoComment}
                        />
                        <Label htmlFor="auto-comment">Auto-comment with AI suggestions</Label>
                      </div>
                      
                      <div className="flex items-center space-x-2">
                        <Switch 
                          id="auto-dm" 
                          checked={autoDM}
                          onCheckedChange={setAutoDM} 
                        />
                        <Label htmlFor="auto-dm">Auto-DM potential leads</Label>
                      </div>
                    </div>
                    
                    <Button onClick={handleSaveAutomation} className="bg-gradient-to-r from-primary to-purple-500 hover:from-primary/90 hover:to-purple-500/90">
                      Save Automation Settings
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
            
            <TabsContent value="notifications" className="space-y-4">
              <Card className="border-primary/10 shadow-md">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Bell className="h-5 w-5" />
                    Notification Preferences
                  </CardTitle>
                  <CardDescription>Control when and how you receive notifications.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div className="space-y-4">
                    <div className="flex items-center space-x-2">
                      <Switch 
                        id="email-notifications" 
                        checked={emailNotifications}
                        onCheckedChange={setEmailNotifications}
                      />
                      <Label htmlFor="email-notifications">Email Notifications</Label>
                    </div>
                    
                    <div className="flex items-center space-x-2">
                      <Switch 
                        id="push-notifications" 
                        checked={pushNotifications}
                        onCheckedChange={setPushNotifications}
                      />
                      <Label htmlFor="push-notifications">Push Notifications</Label>
                    </div>
                    
                    <div className="space-y-2">
                      <Label>Notification Frequency</Label>
                      <select 
                        className="w-full flex h-10 rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                      >
                        <option>Real-time</option>
                        <option>Hourly digest</option>
                        <option>Daily digest</option>
                        <option>Weekly digest</option>
                      </select>
                    </div>
                    
                    <Button onClick={handleSaveNotifications} className="bg-gradient-to-r from-primary to-purple-500 hover:from-primary/90 hover:to-purple-500/90">
                      Save Notification Settings
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
            
            <TabsContent value="account" className="space-y-4">
              <Card className="border-primary/10 shadow-md">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <User className="h-5 w-5" />
                    Account Settings
                  </CardTitle>
                  <CardDescription>Manage your account details and preferences.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="current-password">Current Password</Label>
                      <Input id="current-password" type="password" />
                    </div>
                    
                    <div className="space-y-2">
                      <Label htmlFor="new-password">New Password</Label>
                      <Input id="new-password" type="password" />
                    </div>
                    
                    <div className="space-y-2">
                      <Label htmlFor="confirm-password">Confirm New Password</Label>
                      <Input id="confirm-password" type="password" />
                    </div>
                    
                    <Button className="bg-gradient-to-r from-primary to-purple-500 hover:from-primary/90 hover:to-purple-500/90">
                      Update Password
                    </Button>
                    
                    <div className="pt-4 border-t">
                      <Button variant="destructive">
                        Delete Account
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </main>
    </div>
  );
}
