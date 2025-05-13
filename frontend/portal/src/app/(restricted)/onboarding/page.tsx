import OnboardingGuard from "../../../components/guard/OnboardingGuard";
import ManinForm from "../../../components/Onboarding/MainForm";
import { AuthLoading } from "../dashboard/layout";

export default function Page() {

  return (
    <OnboardingGuard fallback={<AuthLoading />}>
      <ManinForm />
    </OnboardingGuard>
  )
}
