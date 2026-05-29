import { Button } from "@PressScopeBd/ui/components/button";
import { Input } from "@PressScopeBd/ui/components/input";
import { Label } from "@PressScopeBd/ui/components/label";
import { useForm } from "@tanstack/react-form";
import { useNavigate } from "@tanstack/react-router";
import { Eye, EyeOff, Loader2, Mail, Lock } from "lucide-react";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import { z } from "zod";

import { authClient } from "@/lib/auth-client";

import Loader from "./loader";

/* -------------------------------------------------------------------------- */
/*                                   Schema                                   */
/* -------------------------------------------------------------------------- */

const signInSchema = z.object({
  email: z.string().trim().email("Please enter a valid email address"),

  password: z
    .string()
    .min(8, "Password must be at least 8 characters")
    .max(128, "Password is too long"),
});

type SignInFormValues = z.infer<typeof signInSchema>;

/* -------------------------------------------------------------------------- */
/*                              Reusable Helpers                              */
/* -------------------------------------------------------------------------- */

function ErrorMessage({ error }: { error?: string }) {
  if (!error) return null;

  return (
    <p role="alert" className="text-destructive text-sm font-medium">
      {error}
    </p>
  );
}

/* -------------------------------------------------------------------------- */
/*                                Main Component                              */
/* -------------------------------------------------------------------------- */

export default function SignInForm() {
  const navigate = useNavigate({ from: "/" });

  const { isPending: sessionPending } = authClient.useSession();

  const [showPassword, setShowPassword] = useState(false);

  const defaultValues = useMemo<SignInFormValues>(
    () => ({
      email: "",
      password: "",
    }),
    [],
  );

  const form = useForm({
    defaultValues,

    validators: {
      onChange: signInSchema,
      onSubmit: signInSchema,
    },

    onSubmit: async ({ value }) => {
      try {
        await authClient.signIn.email(
          {
            email: value.email,
            password: value.password,
          },
          {
            onSuccess: async () => {
              toast.success("Successfully signed in");

              await navigate({
                to: "/",
              });
            },

            onError: (ctx) => {
              toast.error(
                ctx.error.message ||
                  ctx.error.statusText ||
                  "Failed to sign in",
              );
            },
          },
        );
      } catch (error) {
        console.error(error);

        toast.error("Something went wrong. Please try again.");
      }
    },
  });

  if (sessionPending) {
    return <Loader />;
  }

  return (
    <div className="mx-auto flex h-full w-full max-w-xl flex-col justify-center px-4">
      <div className="space-y-2 text-center">
        <h1 className="text-4xl font-bold tracking-tight">Welcome Back</h1>

        <p className="text-muted-foreground">
          Sign in to continue to your account
        </p>
      </div>

      <form
        className="mt-8 space-y-6"
        noValidate
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();

          void form.handleSubmit();
        }}
      >
        {/* ------------------------------------------------------------------ */}
        {/* Email                                                              */}
        {/* ------------------------------------------------------------------ */}

        <form.Field name="email">
          {(field) => {
            const error = field.state.meta.errors[0]?.message;

            return (
              <div className="space-y-2">
                <Label htmlFor={field.name}>Email Address</Label>

                <div className="relative">
                  <Mail className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />

                  <Input
                    id={field.name}
                    name={field.name}
                    type="email"
                    autoComplete="email"
                    placeholder="you@example.com"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    className="pl-10"
                    aria-invalid={!!error}
                    aria-describedby={`${field.name}-error`}
                  />
                </div>

                <ErrorMessage error={error} />
              </div>
            );
          }}
        </form.Field>

        {/* ------------------------------------------------------------------ */}
        {/* Password                                                           */}
        {/* ------------------------------------------------------------------ */}

        <form.Field name="password">
          {(field) => {
            const error = field.state.meta.errors[0]?.message;

            return (
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor={field.name}>Password</Label>
                </div>

                <div className="relative">
                  <Lock className="text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2" />

                  <Input
                    id={field.name}
                    name={field.name}
                    type={showPassword ? "text" : "password"}
                    autoComplete="current-password"
                    placeholder="Enter your password"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    className="pr-10 pl-10"
                    aria-invalid={!!error}
                    aria-describedby={`${field.name}-error`}
                  />

                  <button
                    type="button"
                    onClick={() => setShowPassword((prev) => !prev)}
                    className="text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2 transition-colors"
                    aria-label={
                      showPassword ? "Hide password" : "Show password"
                    }
                  >
                    {showPassword ? (
                      <EyeOff className="size-4" />
                    ) : (
                      <Eye className="size-4" />
                    )}
                  </button>
                </div>

                <ErrorMessage error={error} />
              </div>
            );
          }}
        </form.Field>

        {/* ------------------------------------------------------------------ */}
        {/* Submit                                                             */}
        {/* ------------------------------------------------------------------ */}

        <form.Subscribe
          selector={(state) => ({
            canSubmit: state.canSubmit,
            isSubmitting: state.isSubmitting,
            isDirty: state.isDirty,
          })}
        >
          {({ canSubmit, isSubmitting, isDirty }) => (
            <Button
              type="submit"
              className="h-11 w-full"
              disabled={!canSubmit || !isDirty || isSubmitting}
            >
              {isSubmitting ? (
                <span className="flex items-center gap-2">
                  <Loader2 className="size-4 animate-spin" />
                  Signing in...
                </span>
              ) : (
                "Sign In"
              )}
            </Button>
          )}
        </form.Subscribe>

        {/* ------------------------------------------------------------------ */}
        {/* Footer                                                             */}
        {/* ------------------------------------------------------------------ */}
      </form>
    </div>
  );
}
