import { useBranches } from "../../../hooks/useBranches";
import { HeroSection } from "../../../components/home/HeroSection";
import { FacilitiesStrip } from "../../../components/home/FacilitiesStrip";
import { BranchesSection } from "../../../components/home/BranchesSection";
import { StayTypeSection } from "../../../components/home/StayTypeSection";
import { CtaSection } from "../../../components/home/CtaSection";

export function HomePage() {
  // ดึงครั้งเดียว ใช้ร่วมกันทั้ง section "สาขาของเรา" และ "ประเภทห้องพัก" (FE-04, FE-05)
  const { activeBranches, loading, error, refetch } = useBranches();

  return (
    <>
      <HeroSection />
      <FacilitiesStrip />
      <BranchesSection activeBranches={activeBranches} loading={loading} error={error} refetch={refetch} />
      <StayTypeSection activeBranches={activeBranches} loading={loading} error={error} refetch={refetch} />
      <CtaSection />
    </>
  );
}
