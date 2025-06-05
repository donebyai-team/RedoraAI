import { FC, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { buildAppUrl } from "@/app/routes";
import { IntegrationType, JWT } from "@doota/pb/doota/portal/v1/portal_pb";
import { errorToMessage } from "@doota/pb/utils/errors";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
import { useIsExecutionRuntimeInPortal } from "@doota/ui-core/hooks/useExecutionRuntime";
import { routes } from "@doota/ui-core/routing";
import { Logo } from "@doota/ui-core/components/Logo";

type Props = {
  onPasswordlessStarted: (message: string) => void
  onPasswordlessVerified: (jwt: JWT) => Promise<void>
  onPasswordlessStartError: (message: string, error: unknown) => void
  onPasswordlessVerifyError: (message: string, error: unknown) => void
}

export const LoginPanel: FC<Props> = ({
  // onPasswordlessStarted,
  onPasswordlessStartError,
  // onPasswordlessVerified,
  // onPasswordlessVerifyError
}) => {
  // const [optState, setOPTState] = useState<'start' | 'verify'>('start')
  // const [email, setEmail] = useState('')
  // const [code, setCode] = useState('')
  const [isLoading, setIsLoading] = useState(false);
  const { portalClient } = useClientsContext()
  const isInPortal = useIsExecutionRuntimeInPortal()

  const handleLoginWithGoogleButton = () => {
    setIsLoading(true);
    portalClient
      .oauthAuthorize({
        integrationType: IntegrationType.GOOGLE,
        redirectUrl: buildAppUrl(routes.app.auth.callback)
      })
      .then(oAuthAuthorizeResp => {
        window.open(oAuthAuthorizeResp.authorizeUrl, '_self')
        setIsLoading(false);
      })
      .catch((err: unknown) => {
        onPasswordlessStartError(errorToMessage(err), err)
        setIsLoading(false);
      })
  }

  let topBoxSx = {}
  if (isInPortal) {
    topBoxSx = { borderWidth: 1, borderStyle: 'solid', borderColor: 'divider' }
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-background to-secondary/20 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <Card className="border-2 border-primary/20 shadow-md">
          {/* Logo Section */}
          <CardHeader className="text-center pb-6">
            <div className='flex w-full justify-center mb-5'>
              <Logo />
            </div>

            <CardTitle className="mt-5 text-3xl font-bold tracking-tight bg-gradient-to-r from-primary to-purple-500 bg-clip-text text-transparent">
              Redora
            </CardTitle>
            <CardDescription className="text-base">
              AI-Powered Reddit Lead Generation
            </CardDescription>
          </CardHeader>

          <CardContent className="space-y-6 px-6">
            {/* Welcome Section */}
            <div className="text-center space-y-2">
              {/* <h2 className="text-xl font-semibold">Welcome Back</h2> */}
              <p className="text-sm text-muted-foreground">
                Sign in to discover your next customers
              </p>
            </div>

            {/* Google Login Button */}
            <Button
              type="button"
              size="lg"
              className="w-full h-12 text-base font-medium bg-white text-gray-900 border border-gray-300 hover:bg-gray-50 shadow-sm"
              onClick={handleLoginWithGoogleButton}
              disabled={isLoading}
            >
              {isLoading ? (
                <div className="flex items-center gap-3">
                  <div className="h-5 w-5 animate-spin rounded-full border-2 border-gray-600 border-r-transparent"></div>
                  <span>Connecting...</span>
                </div>
              ) : (
                <div className="flex items-center gap-3">
                  <svg
                    className="h-5 w-5"
                    viewBox="0 0 24 24"
                    fill="currentColor"
                  >
                    <path
                      d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                      fill="#4285F4"
                    />
                    <path
                      d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                      fill="#34A853"
                    />
                    <path
                      d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                      fill="#FBBC05"
                    />
                    <path
                      d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                      fill="#EA4335"
                    />
                  </svg>
                  <span>Continue with Google</span>
                </div>
              )}
            </Button>

            {/* Feature Highlights */}
            <div className="grid grid-cols-3 gap-4 pt-4">
              <div className="text-center">
                <div className="bg-primary/10 p-3 rounded-lg mx-auto w-fit mb-2">
                  <svg className="h-5 w-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </div>
                <p className="text-xs text-muted-foreground font-medium">AI Powered</p>
              </div>
              <div className="text-center">
                <div className="bg-primary/10 p-3 rounded-lg mx-auto w-fit mb-2">
                  <svg className="h-5 w-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <p className="text-xs text-muted-foreground font-medium">24/7 Scanning</p>
              </div>
              <div className="text-center">
                <div className="bg-primary/10 p-3 rounded-lg mx-auto w-fit mb-2">
                  <svg className="h-5 w-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                  </svg>
                </div>
                <p className="text-xs text-muted-foreground font-medium">Smart DMs</p>
              </div>
            </div>
          </CardContent>

          <CardFooter className="text-center px-6 pb-6">
            <div className="w-full">
              <div className="bg-secondary/30 p-4 rounded-lg border">
                <p className="text-sm text-muted-foreground mb-2">
                  Join thousands of businesses finding their next customers on Reddit
                </p>
                <div className="flex justify-center gap-4 text-xs text-muted-foreground">
                  <span>✓ Secure OAuth</span>
                  <span>✓ Privacy First</span>
                </div>
              </div>

              <p className="text-sm text-muted-foreground mt-4">
                Ready to transform your lead generation?{" "}
                {/* <span className="text-primary font-medium hover:underline cursor-pointer">
                  Watch Demo
                </span> */}
              </p>
            </div>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}