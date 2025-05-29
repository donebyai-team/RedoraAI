import AuthGuard from "../../../components/guard/AuthGuard";
import ManinForm from "../../../components/Onboarding/MainForm";
import { AuthLoading } from "../../old/dashboard/layout";

export default function Page() {

  return (
    <AuthGuard fallback={<AuthLoading />}>
      <ManinForm />
    </AuthGuard>
  );
}
