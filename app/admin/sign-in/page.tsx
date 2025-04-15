import { signInAction } from "@/app/actions";
import { SubmitButton } from "@/components/submit-button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default async function Login() {
  return (
    <form className="flex-1 flex flex-col max-w-md mx-auto">
      <h1 className="text-2xl font-medium">Anmelden</h1>
      <div className="flex flex-col gap-2 [&>input]:mb-3 mt-8">
        <Label htmlFor="email">E-Mail-Adresse</Label>
        <Input name="email" placeholder="beispiel@mail.ch" required />
        <Label htmlFor="password">Passwort</Label>
        <Input
          type="password"
          name="password"
          placeholder="Dein Passwort"
          required
        />
        <SubmitButton pendingText="Anmeldung..." formAction={signInAction}>
          Anmelden
        </SubmitButton>
      </div>
    </form>
  );
}
