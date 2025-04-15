export default async function Layout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="h-[calc(100vh-110px)] overflow-y-auto px-5 pb-5">{children}</div>
  );
}
