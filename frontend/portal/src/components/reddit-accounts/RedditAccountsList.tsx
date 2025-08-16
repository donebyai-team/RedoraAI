"use client";

import { useEffect, useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  ArrowUpRight,
  Cookie,
  ExternalLink,
  Info,
  Key,
  Plus,
  RefreshCw,
} from "lucide-react";
import { Integration, IntegrationState, IntegrationType } from "@doota/pb/doota/portal/v1/portal_pb";
import { portalClient } from "@/services/grpc";
import { getConnectError } from "@/utils/error";
import { FallbackSpinner } from "@/atoms/FallbackSpinner";
import { buildAppUrl } from "@/app/routes";
import { routes } from "@doota/ui-core/routing";
import toast from "react-hot-toast";
import { isPlatformAdmin } from "@doota/ui-core/helper/role";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Textarea } from "@/components/ui/textarea"
import { useAuthUser, useAuth } from "@doota/ui-core/hooks/useAuth";
import Link from "next/link";
import countries from "i18n-iso-countries";
import enLocale from "i18n-iso-countries/langs/en.json";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../ui/select";

countries.registerLocale(enLocale);
const countryEntries = Object.entries(
  countries.getNames("en", { select: "official" })
);

interface RedditAccount {
  id: string;
  username: string;
  countryCode?: string;
  karma: number;
  apiIntegration?: Integration;
  cookieIntegration?: Integration;
}

export function RedditAccountsList() {
  // Mock data - would come from props or API in real implementation
  const [accounts, setAccounts] = useState<RedditAccount[]>([]);
  const [loading, setLoading] = useState(true)
  const user = useAuthUser()
  const { setUser, setOrganization } = useAuth()
  const [cookieCountry, setCookieCountry] = useState<string | null>(null);

  const fetchAccounts = () => {
    setLoading(true);
    portalClient.getIntegrations({})
      .then((res) => {
        const arr: RedditAccount[] = [];

        res.integrations.forEach((integration: Integration) => {
          const username = integration.details.value?.userName;
          if (!username) return;

          // Find if we already have this username in the array
          let account = arr.find(acc => acc.username === username);
          if (!account) {
            account = {
              id: username,
              username,
              karma: 0,
            };
            arr.push(account);
          }

          // Fill in API integration
          if (integration.type === IntegrationType.REDDIT) {
            account.apiIntegration = integration
          }

          // Fill in Cookie integration
          if (integration.type === IntegrationType.REDDIT_DM_LOGIN) {
            account.countryCode = integration.details.value?.alpha2CountryCode
            account.cookieIntegration = integration
          }
        });

        setAccounts(arr);
      })
      .catch((err) => {
        toast.error(getConnectError(err));
      })
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchAccounts();
  }, []);

  const openOauthConsentScreen = (integrationType: IntegrationType) => {
    portalClient
      .oauthAuthorize({
        integrationType: integrationType,
        redirectUrl: buildAppUrl(routes.new.integrations),
      })
      .then(oAuthAuthorizeResp => {
        window.open(oAuthAuthorizeResp.authorizeUrl, '_self')
      })
  }

  const handleCountryChange = (id: string, countryCode: string) => {
    console.log("country code changed", countryCode, "id", id);

    setAccounts(prev =>
      prev.map(account =>
        account.id === id ? { ...account, countryCode } : account
      )
    );

    // Run async update separately
    const account = accounts.find(a => a.id === id);
    if (account?.cookieIntegration) {
      portalClient.updateIntegration({
        id: account.cookieIntegration.id,
        details: {
          case: "reddit",
          value: {
            alpha2CountryCode: countryCode,
          },
        },
      }).then(() => {
        console.log("Country updated");
        toast.success("Integration updated!")
      }).catch(err => {
        toast.error(getConnectError(err));
      });
    }
  };


  const handleDisconnectReddit = (id: string) => {
    // Optimistic update
    setAccounts(prev =>
      prev.map(account => {
        if (account.apiIntegration?.id === id) {
          return {
            ...account,
            apiIntegration: {
              ...account.apiIntegration,
              status: IntegrationState.AUTH_REVOKED
            }
          };
        }
        if (account.cookieIntegration?.id === id) {
          return {
            ...account,
            cookieIntegration: {
              ...account.cookieIntegration,
              status: IntegrationState.AUTH_REVOKED
            }
          };
        }
        return account;
      })
    );

    // API call
    portalClient.revokeIntegration({ id })
      .then(() => {
        setAccounts(prev => {
          const updated = prev.map(account => {
            if (account.apiIntegration?.id === id) {
              return {
                ...account,
                apiIntegration: {
                  ...account.apiIntegration,
                  status: IntegrationState.AUTH_REVOKED
                }
              };
            }
            if (account.cookieIntegration?.id === id) {
              return {
                ...account,
                cookieIntegration: {
                  ...account.cookieIntegration,
                  status: IntegrationState.AUTH_REVOKED
                }
              };
            }
            return account;
          });

          // Check for active API integrations
          const hasActiveApiIntegration = updated.some(
            acc => acc.apiIntegration?.status === IntegrationState.ACTIVE
          );

          // Check for active Cookie integrations
          const hasActiveCookieIntegration = updated.some(
            acc => acc.cookieIntegration?.status === IntegrationState.ACTIVE
          );

          if (!hasActiveApiIntegration) {
            console.log("No active API integration, disabling automated comments");
            handleSaveAutomation({ comment: { enabled: false } });
          }

          if (!hasActiveCookieIntegration) {
            console.log("No active cookie integration, disabling automated DMs");
            handleSaveAutomation({ dm: { enabled: false } });
          }

          return updated;
        });
      })
      .catch(err => {
        toast.error(getConnectError(err));
      });
  };

  const [showCookieModal, setShowCookieModal] = useState(false);
  const [cookieInput, setCookieInput] = useState('');
  const [cookieError, setCookieError] = useState('');
  const [isSubmittingCookie, setIsSubmittingCookie] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);


  const handleConnectReddit = async (id: string) => {
    const countryCode = accounts.find(acc => acc.id === id)?.countryCode;
    if (!countryCode) {
      toast.error("Please select a country before continuing");
      return;
    }


    let popup: Window | null = null;
    try {
      setIsConnecting(true);
      popup = window.open('', '_blank', "width=600,height=800");
      if (!popup) {
        toast.error('Popup was blocked. Please allow popups in your browser.');
        return;
      }
      // Inject temporary loading UI
      popup.document.write(`
                  <html>
                      <head><title>Connecting...</title></head>
                      <body style="display:flex;justify-content:center;align-items:center;height:100vh;font-family:sans-serif;">
                  <div>
                      <p>Connecting to Reddit Chat...</p>
                  </div>
                  </body>
                  </html>
              `);
      popup.document.close();
      const abortController = new AbortController();
      const response = portalClient.connectReddit({ alpha2CountryCode: countryCode }, { signal: abortController.signal });

      let streamClosed = false;

      // Set interval to check if popup closed manually
      const popupCheckInterval = setInterval(() => {
        if (popup && popup.closed && !streamClosed) {
          setIsConnecting(false);
          clearInterval(popupCheckInterval);
          streamClosed = true;
          abortController.abort(); // ⛔ cancels the stream
        }
      }, 500); // check every 500ms

      for await (const msg of response) {
        if (msg.url) {
          // Open the Reddit login page in a popup
          popup.location.href = msg.url; // Redirect once URL is available
        }
      }

      // Stream finished normally
      streamClosed = true;
      clearInterval(popupCheckInterval);
      // Stream has ended successfully
      if (popup && !popup.closed) {
        popup.close();
      }

      await handleSaveAutomation({ dm: { enabled: true } });

      // Refresh integrations to reflect the newly connected status
      fetchAccounts()

      toast.success("Reddit connected successfully");
    } catch (err: any) {
      if (popup && !popup.closed) {
        popup.close();
      }
      setCookieCountry(countryCode);
      setShowCookieModal(true); // show modal
      // toast.error(getConnectError(err));
    } finally {
      setIsConnecting(false);
    }
  };


  const handleSaveAutomation = async (req: any) => {
    try {
      console.log("Updating autmation", req);
      const result = await portalClient.updateAutomationSettings(req);

      if (isPlatformAdmin(user)) {
        setOrganization(result);
      }

      setUser(prev => {
        if (!prev) return prev
        const updatedOrganizations = prev.organizations.map(org =>
          org.id === result.id ? result : org
        )
        return { ...prev, organizations: updatedOrganizations }
      })
    } catch (err) {
      toast.error(getConnectError(err));
    }
  }



  const getIntegrationBadge = (integration?: Integration) => {
    const baseClass =
      "flex items-center gap-1 border-green-200 bg-green-50 text-green-700";

    if (!integration) {
      return (
        <Badge variant="outline" className={baseClass}>
          <Cookie className="h-3 w-3" />
          Connect Cookies
        </Badge>
      );
    }

    if (integration.status === IntegrationState.ACTIVE) {
      return (
        <Badge variant="outline" className={baseClass}>
          Connected
        </Badge>
      );
    }

    // For inactive integrations
    return (
      <Badge variant="outline" className={baseClass}>
        <Cookie className="h-3 w-3" />
        Reconnect
      </Badge>
    );
  };


  if (loading) {
    return <FallbackSpinner />
  }

  return (
    <Card className="border-primary/10 shadow-md p-5">
      <CardHeader>
        <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
          <div>
            <CardTitle className="flex items-center gap-2">
              Connected Reddit Accounts
            </CardTitle>
            <CardDescription>
              <div
                style={{
                  padding: "12px",
                  marginTop: "8px",
                  borderLeft: "4px solid #ff4500",
                  backgroundColor: "#fff8f6",
                }}
              >
                <p style={{ color: "#4d2c19", margin: 0 }}>
                  <strong>Note:</strong> You can connect multiple Reddit accounts. We will
                  automatically rotate between them when sending comment and DMs.
                </p>
              </div>

            </CardDescription>
          </div>
          <Button
            onClick={() => openOauthConsentScreen(IntegrationType.REDDIT)}
            className="gap-1">
            <Plus className="h-4 w-4" />
            Add Reddit Account
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {accounts.length === 0 ? (
          <div className="text-center py-8 border border-dashed rounded-md">
            <p className="text-muted-foreground mb-4">
              No Reddit accounts connected yet. Add an account to start engaging with leads.
            </p>
            <Button
              onClick={() => openOauthConsentScreen(IntegrationType.REDDIT)}
              className="gap-1"
            >
              <Plus className="h-4 w-4" />
              Connect Account
            </Button>
          </div>
        ) : (
          <div className="space-y-6">
            {accounts.map((account) => (
              <Card key={account.id} className="border border-border/50">
                <CardHeader className="pb-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-full bg-secondary/80 flex items-center justify-center">
                        {account.username.charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-medium">u/{account.username}</span>
                          <a
                            href={`https://reddit.com/user/${account.username}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-muted-foreground hover:text-primary"
                          >
                            <ArrowUpRight className="h-3 w-3" />
                          </a>
                        </div>
                        {/* <span className="text-xs text-muted-foreground">
                          {account.karma.toLocaleString()} karma
                        </span> */}
                      </div>
                    </div>

                    <div className="flex items-center gap-2">
                      {/* Account actions removed - manage integrations individually */}
                    </div>
                  </div>
                </CardHeader>

                <CardContent className="space-y-4 pt-0">
                  {/* API Integration Section */}
                  {account.apiIntegration?.id && (
                    <div className="p-3 border rounded-lg bg-background">
                      <div className="flex items-center justify-between mb-2">
                        <div className="flex items-center gap-2">
                          <Key className="h-4 w-4 text-muted-foreground" />
                          <span className="font-medium text-sm">API Integration</span>
                          {getIntegrationBadge(account.apiIntegration)}
                        </div>
                        <div className="flex items-center gap-2">
                          {account.apiIntegration.status !== IntegrationState.ACTIVE ? (
                            <Button
                              onClick={() => openOauthConsentScreen(IntegrationType.REDDIT)}
                              variant="outline"
                              size="sm"
                              className="gap-1"
                            >
                              <RefreshCw className="h-3 w-3" />
                              Reconnect
                            </Button>
                          ) : (
                            <Button
                              onClick={() => handleDisconnectReddit(account.apiIntegration!.id)}
                              variant="outline"
                              size="sm"
                              className="gap-1 text-destructive border-destructive/20 hover:bg-destructive/10">
                              Revoke
                            </Button>
                          )}
                        </div>
                      </div>

                      <p className="text-xs text-red-600 mb-2">
                        {account.apiIntegration.details.value?.reason}
                      </p>

                      <p className="text-xs text-muted-foreground">
                        Used for posting comments and basic Reddit interactions
                      </p>
                    </div>
                  )}

                  {/* Cookie Integration Section */}
                  <div className="p-3 border rounded-lg bg-background">
                    <Dialog open={showCookieModal} onOpenChange={setShowCookieModal}>
                      <DialogContent className="max-w-lg">
                        <DialogHeader>
                          <DialogTitle>Connect Reddit Account Manually</DialogTitle>
                        </DialogHeader>

                        <Card className="bg-gray-50 rounded-lg p-4 mb-5">
                          <CardContent className="p-0 space-y-3">
                            <p className="text-sm">
                              We couldn't connect to Reddit automatically. Please follow the steps below to connect your account manually:
                            </p>
                            <ol className="list-decimal pl-4 space-y-1 text-sm">
                              <li>
                                Go to{" "}
                                <Link
                                  href="https://www.reddit.com"
                                  target="_blank"
                                  rel="noopener"
                                  className="underline text-blue-600"
                                >
                                  reddit.com
                                </Link>{" "}
                                and log in to your Reddit account.
                              </li>
                              <li>
                                Install the Chrome extension{" "}
                                <Link
                                  href="https://chromewebstore.google.com/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm"
                                  target="_blank"
                                  rel="noopener"
                                  className="underline text-blue-600"
                                >
                                  EditThisCookie
                                </Link>.
                              </li>
                              <li>Open the extension and copy all cookies for reddit.com in JSON format.</li>
                              <li>
                                Paste the copied cookie JSON into the field below and click <strong>Submit</strong>.
                              </li>
                            </ol>
                          </CardContent>
                        </Card>

                        <div className="space-y-2">
                          <Textarea
                            rows={5}
                            placeholder="Paste your Reddit cookies JSON here..."
                            value={cookieInput}
                            onChange={(e) => setCookieInput(e.target.value)}
                            className={cookieError ? "border-red-500" : ""}
                          />
                          {cookieError ? (
                            <p className="text-xs text-red-600">{cookieError}</p>
                          ) : (
                            <p className="text-xs text-gray-500">
                              Paste your cookies in JSON format. Validation will take a few minutes.
                            </p>
                          )}
                        </div>

                        <DialogFooter className="mt-4">
                          <Button
                            variant="outline"
                            onClick={() => setShowCookieModal(false)}
                            disabled={isSubmittingCookie}
                          >
                            Cancel
                          </Button>
                          <Button
                            onClick={async () => {
                              try {
                                if (!cookieCountry) {
                                  setCookieError("Please select a country before continuing");
                                  return;
                                }
                                setIsSubmittingCookie(true)
                                setCookieError("")

                                // ✅ Validate JSON format
                                try {
                                  JSON.parse(cookieInput)
                                } catch {
                                  setCookieError("Please enter valid JSON format.")
                                  return
                                }

                                const response = portalClient.connectReddit({
                                  cookieJson: cookieInput,
                                  alpha2CountryCode: cookieCountry,
                                })
                                for await (const msg of response) { }

                                await handleSaveAutomation({ dm: { enabled: true } })
                                fetchAccounts()
                                setShowCookieModal(false)
                                setCookieInput("")
                              } catch (e: any) {
                                setCookieError(getConnectError(e))
                              } finally {
                                setIsSubmittingCookie(false)
                              }
                            }}
                            disabled={isSubmittingCookie}
                          >
                            {isSubmittingCookie ? "Submitting..." : "Submit"}
                          </Button>
                        </DialogFooter>
                      </DialogContent>
                    </Dialog>
                    <div className="flex items-center justify-between mb-3">
                      <div className="flex items-center gap-2">
                        <Cookie className="h-4 w-4 text-muted-foreground" />
                        <span className="font-medium text-sm">Cookie Integration</span>
                        {account.cookieIntegration && getIntegrationBadge(account.cookieIntegration)}
                      </div>
                      <div className="flex items-center gap-2">
                        {!account.cookieIntegration || account.cookieIntegration.status !== IntegrationState.ACTIVE ? (
                          <Button
                            onClick={() => handleConnectReddit(account.id)}
                            size="sm"
                            className="gap-1 bg-orange-500 text-white hover:bg-orange-600 shadow-md hover:shadow-lg transition-all duration-200 border-none"
                          >
                            {account.cookieIntegration ? "Reconnect" : "Connect"}
                          </Button>

                        ) : (
                          <Button
                            onClick={() => handleDisconnectReddit(account.cookieIntegration!.id)}
                            variant="outline" size="sm" className="gap-1 text-destructive border-destructive/20 hover:bg-destructive/10">
                            Revoke
                          </Button>
                        )}
                      </div>
                    </div>

                    {account.cookieIntegration && (
                      <p className="text-xs text-red-600 mb-3">
                        {account.cookieIntegration.details.value?.reason}
                      </p>
                    )}

                    {/* Country Selection */}
                    <div className="space-y-2 mb-3">
                      <div className="flex items-center gap-2">
                        <label className="text-xs font-medium text-foreground">Country:</label>
                        <Select
                          value={account.countryCode || ""}
                          onValueChange={(value) => handleCountryChange(account.id, value)}
                        >
                          <SelectTrigger className="w-48 h-7 text-xs">
                            <SelectValue placeholder="Select a country" />
                          </SelectTrigger>
                          <SelectContent>
                            {countryEntries.map(([code, name]) => (
                              <SelectItem key={code} value={code} className="text-xs">
                                {name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                      <p className="text-xs text-muted-foreground">
                        Select the country you primarily use this Reddit account from. We’ll use it to match your proxy location and help avoid Reddit security checks.
                      </p>
                    </div>

                    {(!account.cookieIntegration || account.cookieIntegration?.status !== IntegrationState.ACTIVE) ? (
                      <div className="space-y-3">
                        <div className="bg-blue-50 border border-blue-200 rounded-md p-3">
                          <div className="flex items-start gap-2">
                            <Info className="h-4 w-4 text-blue-600 mt-0.5 flex-shrink-0" />
                            <div className="text-xs text-blue-800">
                              <p className="font-medium mb-1">Why connect cookies?</p>
                              <p>Enable advanced features like sending direct messages and automated warm activities to simulate real user behavior.</p>
                            </div>
                          </div>
                        </div>

                        <div className="space-y-2">
                          <h4 className="text-xs font-medium text-foreground">Setup Instructions:</h4>
                          <ul className="text-xs text-muted-foreground space-y-1 list-disc pl-4">
                            <li>Your credentials are never stored — we use browser cookies to simulate real user behavior</li>
                            <li>Log in using your Reddit email (or username) and password</li>
                            <li>If your account doesn&apos;t have a password set, you&apos;ll need to create one first</li>
                          </ul>

                          <div className="pt-2">
                            <a
                              href="https://redoraai.featurebase.app/en/help/articles/9204295-how-to-enable-dm-automation"
                              target="_blank"
                              rel="noopener noreferrer"
                              className="inline-flex items-center gap-1 text-xs text-primary hover:underline"
                            >
                              How do I add a password to my account? — Reddit Help
                              <ExternalLink className="h-3 w-3" />
                            </a>
                          </div>
                        </div>
                      </div>
                    ) : (
                      <p className="text-xs text-muted-foreground">
                        Used for sending DMs and perform automated activities like account warmups.
                      </p>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}