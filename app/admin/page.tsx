import { createClient } from "@/utils/supabase/server";

export default async function HomePage() {
  const supabase = await createClient();

  const {
    data: { user },
  } = await supabase.auth.getUser();

  return (
    <div className="flex-1 flex flex-col gap-12 max-w-md mx-auto">
      <div className="flex flex-col gap-2 items-start w-full">
        <h2 className="font-bold text-2xl mb-4">Your user details</h2>
        <pre className="text-xs font-mono p-3 rounded border max-h-32 overflow-auto w-full">
          {JSON.stringify(user, null, 2)}
        </pre>
      </div>
    </div>
  );
}
