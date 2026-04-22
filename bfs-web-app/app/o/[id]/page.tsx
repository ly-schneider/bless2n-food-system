import { redirect } from "next/navigation"

export default async function ShortOrderRedirect({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  redirect(`/food/orders/${encodeURIComponent(id)}`)
}
