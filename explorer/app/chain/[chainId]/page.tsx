import ChainPage from "@/components/chain-page";

export default async function Page({ params }: { params: Promise<{ chainId: string }> }) {
  const { chainId } = await params;
  return <ChainPage chainId={chainId} />;
}
