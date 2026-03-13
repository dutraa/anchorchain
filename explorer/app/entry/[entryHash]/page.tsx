import EntryPage from "@/components/entry-page";

export default async function Page({ params }: { params: Promise<{ entryHash: string }> }) {
  const { entryHash } = await params;
  return <EntryPage entryHash={entryHash} />;
}
