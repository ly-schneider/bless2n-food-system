"use client";

import { signInAction } from "@/app/actions";
import { SubmitButton } from "@/components/submit-button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";

function ErrorMessage() {
  const searchParams = useSearchParams();
  const error = searchParams.get("error");
  const errorMessage =
    error === "Invalid login credentials" ? (
      <p className="text-red-500 text-sm">
        Ungültige E-Mail-Adresse oder Passwort.
      </p>
    ) : (
      <p className="text-red-500 text-sm">
        Fehler bei der Anmeldung. Bitte versuche es erneut.
      </p>
    );

  return error && errorMessage;
}

export default function Login() {
  const [showPassword, setShowPassword] = useState(false);

  const togglePasswordVisibility = () => {
    setShowPassword((prev) => !prev);
  };

  return (
    <form className="flex-1 flex flex-col max-w-md mx-auto">
      <h1 className="text-2xl font-medium">Anmelden</h1>
      <div className="flex flex-col gap-2 [&>input]:mb-3 mt-8">
        <Label htmlFor="email">E-Mail-Adresse</Label>
        <Input
          name="email"
          placeholder="beispiel@mail.ch"
          required
          type="email"
        />
        <Label htmlFor="password">Passwort</Label>
        <Input
          type={showPassword ? "text" : "password"}
          name="password"
          placeholder="Dein Passwort"
          required
        />
        <button
          type="button"
          className="text-sm text-gray-500 hover:text-gray-700 mb-3"
          onClick={togglePasswordVisibility}
        >
          {showPassword ? "Passwort ausblenden" : "Passwort anzeigen"}
        </button>
        <SubmitButton pendingText="Anmeldung..." formAction={signInAction}>
          Anmelden
        </SubmitButton>
        <Suspense>
          <ErrorMessage />
        </Suspense>
      </div>
    </form>
  );
}
