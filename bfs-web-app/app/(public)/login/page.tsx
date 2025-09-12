import { ArrowLeft, EyeOff } from "lucide-react"
import { Metadata } from "next"
import Link from "next/link"
import { redirect } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Separator } from "@/components/ui/separator"

export const metadata: Metadata = {
  title: "Sign In - Bless2n Food System",
  description: "Sign in to your account or continue as guest.",
}

export default function LoginPage({
  searchParams,
}: {
  searchParams: { [key: string]: string | string[] | undefined }
}) {
  const redirect = searchParams.redirect as string
  const error = searchParams.error as string

  return (
    <div className="from-primary/5 to-secondary/5 flex min-h-screen items-center justify-center bg-gradient-to-br px-4">
      <div className="w-full max-w-md space-y-6">
        {/* Back to home */}
        <div className="flex items-center">
          <Button variant="ghost" size="sm" asChild>
            <Link href="/">
              <ArrowLeft className="mr-2 h-4 w-4" />
              Back to Home
            </Link>
          </Button>
        </div>

        {/* Sign In Card */}
        <Card>
          <CardHeader className="text-center">
            <CardTitle className="text-2xl">Welcome Back</CardTitle>
            <CardDescription>Sign in to your account to access your orders and saved preferences</CardDescription>

            {error && (
              <div className="bg-destructive/10 border-destructive/20 mt-4 rounded-md border p-3">
                <p className="text-destructive text-sm">
                  {error === "admin_required" && "Admin access required to view this page"}
                  {error === "pos_required" && "POS access required to view this page"}
                  {error === "unauthorized" && "You are not authorized to access this page"}
                  {!["admin_required", "pos_required", "unauthorized"].includes(error) && "Invalid credentials"}
                </p>
              </div>
            )}
          </CardHeader>

          <CardContent className="space-y-4">
            <LoginForm redirect={redirect} />

            <div className="relative">
              <div className="absolute inset-0 flex items-center">
                <Separator />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-background text-muted-foreground px-2">Or</span>
              </div>
            </div>

            <Button variant="outline" className="w-full" asChild>
              <Link href={redirect ? `/?redirect=${encodeURIComponent(redirect)}` : "/"}>Continue as Guest</Link>
            </Button>
          </CardContent>

          <CardFooter className="text-center">
            <p className="text-muted-foreground text-sm">
              Don't have an account?{" "}
              <Link href="/register" className="text-primary font-medium hover:underline">
                Sign up here
              </Link>
            </p>
          </CardFooter>
        </Card>

        {/* Guest Benefits */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Why create an account?</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="text-muted-foreground space-y-2 text-sm">
              <li className="flex items-center">
                <span className="bg-primary mr-3 h-2 w-2 rounded-full" />
                Save your favorite orders
              </li>
              <li className="flex items-center">
                <span className="bg-primary mr-3 h-2 w-2 rounded-full" />
                Track order history
              </li>
              <li className="flex items-center">
                <span className="bg-primary mr-3 h-2 w-2 rounded-full" />
                Faster checkout
              </li>
              <li className="flex items-center">
                <span className="bg-primary mr-3 h-2 w-2 rounded-full" />
                Exclusive offers and rewards
              </li>
            </ul>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function LoginForm({ redirect: _redirect }: { redirect?: string }) {
  const handleSubmit = async (formData: FormData) => {
    "use server"

    const email = formData.get("email") as string
    const password = formData.get("password") as string

    // Real login via AuthService
    try {
      const { AuthService } = await import("@/lib/auth")
      // Map password input to OTP field for current backend
      await AuthService.login({ email, otp: password })
      redirect(_redirect || "/")
    } catch (err) {
      console.error("Login failed", err)
      redirect(`/login?error=unauthorized${_redirect ? `&redirect=${encodeURIComponent(_redirect)}` : ""}`)
    }
  }

  return (
    <form action={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="email">Email</Label>
        <Input id="email" name="email" type="email" placeholder="Enter your email" required autoComplete="email" />
      </div>

      <div className="space-y-2">
        <Label htmlFor="password">Password</Label>
        <div className="relative">
          <Input
            id="password"
            name="password"
            type="password"
            placeholder="Enter your password"
            required
            autoComplete="current-password"
            className="pr-10"
          />
          <button
            type="button"
            className="text-muted-foreground hover:text-foreground absolute inset-y-0 right-0 flex items-center pr-3"
          >
            <EyeOff className="h-4 w-4" />
          </button>
        </div>
      </div>

      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <input id="remember" name="remember" type="checkbox" className="rounded border-gray-300" />
          <Label htmlFor="remember" className="text-sm">
            Remember me
          </Label>
        </div>

        <Link href="/forgot-password" className="text-primary text-sm font-medium hover:underline">
          Forgot password?
        </Link>
      </div>

      <Button type="submit" className="w-full">
        Sign In
      </Button>
    </form>
  )
}
