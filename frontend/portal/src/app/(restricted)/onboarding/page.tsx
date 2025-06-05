import Onboarding from "@/components/pages/Onboarding";
import AuthGuard from "@/components/guard/AuthGuard";
import { AuthLoading } from "../../old/dashboard/layout";

export default function Page() {

  return (
    <AuthGuard fallback={<AuthLoading />}>
      <Onboarding />
    </AuthGuard>
  );
}
