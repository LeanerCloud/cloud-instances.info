"use client";

import InstanceTable from "@/components/InstanceTable";
import Filters from "@/components/Filters";
import DoMigration from "@/components/DoMigration";
import MarketingWrapper from "@/components/MarketingWrapper";
import { useState } from "react";

interface GCPClientProps {
  columnAtomKey: string;
  regions: any;
  instances: any[];
  instanceCount: number;
}

export function GCPClient({
  columnAtomKey,
  regions,
  instances,
  instanceCount,
}: GCPClientProps) {
  const [rowSelection, setRowSelection] = useState({});
  
  // Convert regions array to the expected format
  const formattedRegions = {
    main: regions.reduce((acc: any, region: any) => {
      acc[region.name] = region.displayname || region.name;
      return acc;
    }, {}),
    local_zone: {},
    wavelength: {}
  };

  return (
    <MarketingWrapper>
      <div className="flex flex-col h-[calc(100vh-4rem)] overflow-hidden">
        <DoMigration />
        <Filters 
          columnAtomKey={columnAtomKey}
          regions={formattedRegions}
          rowSelection={rowSelection}
          reservedTermOptions={[
            { value: "1yr", label: "1 year" },
            { value: "3yr", label: "3 years" }
          ]}
        />
        <InstanceTable
          instances={instances}
          columnAtomKey={columnAtomKey}
          regions={regions}
          rowSelection={rowSelection}
          setRowSelection={setRowSelection}
          instanceCount={instanceCount}
        />
      </div>
    </MarketingWrapper>
  );
}