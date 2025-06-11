import { FC, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { buildAppUrl } from "@/app/routes";
import { IntegrationType, JWT } from "@doota/pb/doota/portal/v1/portal_pb";
import { errorToMessage } from "@doota/pb/utils/errors";
import { useClientsContext } from "@doota/ui-core/context/ClientContext";
// import { useIsExecutionRuntimeInPortal } from "@doota/ui-core/hooks/useExecutionRuntime";
import { routes } from "@doota/ui-core/routing";
import { Logo } from "@doota/ui-core/components/Logo";
import { toast } from "@/hooks/use-toast";
import { Input } from "../ui/input";
import { InputOTP, InputOTPGroup, InputOTPSlot } from "../ui/input-otp";
import { useRouter } from "next/navigation";

type Props = {
  onPasswordlessStarted: (message: string) => void
  onPasswordlessVerified: (jwt: JWT) => Promise<void>
  onPasswordlessStartError: (message: string, error: unknown) => void
  onPasswordlessVerifyError: (message: string, error: unknown) => void
}

export const LoginPanel: FC<Props> = ({
  // onPasswordlessStarted,
  onPasswordlessStartError,
  onPasswordlessVerified,
  // onPasswordlessVerifyError
}) => {
  // const [optState, setOPTState] = useState<'start' | 'verify'>('start')
  // const [email, setEmail] = useState('')
  // const [code, setCode] = useState('')
  const [isLoading, setIsLoading] = useState(false);
  const { portalClient } = useClientsContext()

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

  const [email, setEmail] = useState("");
  const [otp, setOtp] = useState("");
  const [showOtp, setShowOtp] = useState(false);
  const [isEmailLoading, setIsEmailLoading] = useState(false);
  const [emailError, setEmailError] = useState("");
  const navigate = useRouter();

  const validateEmail = (email: string) => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  const handleEmailSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email) {
      setEmailError("Email is required");
      return;
    }

    if (!validateEmail(email)) {
      setEmailError("Please enter a valid email address");
      return;
    }

    setEmailError("");
    setIsEmailLoading(true);

    try {
      // Simulate sending OTP to email
      await portalClient.passwordlessStart({ email: email });
      
      setShowOtp(true);
      toast({
        title: "OTP sent",
        description: `Verification code sent to ${email}`,
      });
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to send verification code. Please try again.",
        variant: "destructive",
      });
    } finally {
      setIsEmailLoading(false);
    }
  };

  const handleOtpSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (otp.length !== 6) return;

    setIsEmailLoading(true);

    try {
      // Simulate OTP verification
      await new Promise(resolve => setTimeout(resolve, 1000));

      const token = await portalClient.passwordlessVerify({ email: email, code: otp });
      onPasswordlessVerified(token);
      
      navigate.push("/dashboard");
      
    } catch (error) {
      toast({
        title: "Invalid code",
        description: "Please check your verification code and try again.",
        variant: "destructive",
      });
    } finally {
      setIsEmailLoading(false);
    }
  };

  const handleBackToEmail = () => {
    setShowOtp(false);
    setOtp("");
  };

  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
    if (emailError) {
      setEmailError("");
    }
  };


  return (
    <div className="min-h-screen bg-gradient-to-b from-background to-secondary/20 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <Card className="border-2 border-primary/20 shadow-md">
          {/* Logo Section */}
          <CardHeader className="text-center pb-6">
            <div className="flex justify-center mb-4">
              <Logo />
            </div>

            <CardTitle className="text-3xl font-bold tracking-tight bg-gradient-to-r from-primary to-purple-500 bg-clip-text text-transparent">
              Redora
            </CardTitle>
            <CardDescription className="text-base">
              Sign in to discover your next customers
            </CardDescription>
          </CardHeader>

          <CardContent className="space-y-6 px-6">
            {/* Email Login Section */}
            {!showOtp ? (
              <form onSubmit={handleEmailSubmit} className="space-y-4">
                <div className="space-y-2">
                  <Input
                    id="email"
                    type="email"
                    placeholder="Enter your email"
                    value={email}
                    onChange={handleEmailChange}
                    required
                    className={emailError ? "border-destructive" : ""}
                  />
                  {emailError && (
                    <p className="text-sm text-destructive">{emailError}</p>
                  )}
                </div>
                <Button
                  type="submit"
                  size="lg"
                  className="w-full h-12 text-base font-medium"
                  disabled={isEmailLoading || !email}
                >
                  {isEmailLoading ? (
                    <div className="flex items-center gap-3">
                      <div className="h-5 w-5 animate-spin rounded-full border-2 border-white border-r-transparent"></div>
                      <span>Sending code...</span>
                    </div>
                  ) : (
                    "Continue with Email"
                  )}
                </Button>
              </form>
            ) : (
              <form onSubmit={handleOtpSubmit} className="space-y-4">
                <div className="space-y-2">
                  <label htmlFor="otp" className="text-sm font-medium">
                    Verification Code
                  </label>
                  <p className="text-sm text-muted-foreground">
                    Enter the 6-digit code sent to {email}
                  </p>
                  <div className="flex justify-center">
                    <InputOTP
                      value={otp}
                      onChange={(value) => setOtp(value)}
                      maxLength={6}
                    >
                      <InputOTPGroup>
                        <InputOTPSlot index={0} />
                        <InputOTPSlot index={1} />
                        <InputOTPSlot index={2} />
                        <InputOTPSlot index={3} />
                        <InputOTPSlot index={4} />
                        <InputOTPSlot index={5} />
                      </InputOTPGroup>
                    </InputOTP>
                  </div>
                </div>
                <Button
                  type="submit"
                  size="lg"
                  className="w-full h-12 text-base font-medium"
                  disabled={isEmailLoading || otp.length !== 6}
                >
                  {isEmailLoading ? (
                    <div className="flex items-center gap-3">
                      <div className="h-5 w-5 animate-spin rounded-full border-2 border-white border-r-transparent"></div>
                      <span>Verifying...</span>
                    </div>
                  ) : (
                    "Verify & Sign In"
                  )}
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="lg"
                  className="w-full"
                  onClick={handleBackToEmail}
                >
                  Back to Email
                </Button>
              </form>
            )}

            {/* Divider */}
            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-background px-2 text-muted-foreground">
                  Or continue with
                </span>
              </div>
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
              </p>
            </div>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}