"use client";

import InstanceTable from "@/components/InstanceTable";
import Filters from "@/components/Filters";
import DoMigration from "@/components/DoMigration";
import MarketingWrapper from "@/components/MarketingWrapper";
import { useInstanceData } from "@/state";
import { GCPInstance } from "@/utils/colunnData/gcp";
import dynamicallyDecompress from "@/utils/dynamicallyDecompress";
import { RowSelectionState } from "@tanstack/react-table";
import { useMemo, useState } from "react";

interface GCPClientProps {
  columnAtomKey: "gcp";
  regions: any;
  instanceCount: number;
  compressedDataPathTemplate: string;
  compressedInstances: [string[], ...GCPInstance[]];
}

export function GCPClient({
  columnAtomKey,
  regions,
  instanceCount,
  compressedDataPathTemplate,
  compressedInstances,
}: GCPClientProps) {
  const initialInstances = useMemo(() => {
    const rainbowTable = compressedInstances.shift() as string[];
    if (!Array.isArray(rainbowTable)) {
      // This is probably dev.
      compressedInstances.unshift(rainbowTable);
      return compressedInstances as GCPInstance[];
    }
    return compressedInstances.map((instance) =>
      dynamicallyDecompress(instance as GCPInstance, rainbowTable),
    );
  }, [compressedInstances]);

  const allInstances = useInstanceData(
    compressedDataPathTemplate,
    initialInstances,
  );
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
  
  // Convert regions array to the expected format
  const formattedRegions = {
    main: Array.isArray(regions) 
      ? regions.reduce((acc: any, region: any) => {
          acc[region.name] = region.displayname || region.name;
          return acc;
        }, {})
      : regions,
    local_zone: {},
    wavelength: {}
  };

  return (
    <MarketingWrapper>
      <div className="flex flex-col h-[calc(100vh-4rem)] overflow-hidden">
        <DoMigration atomKey={columnAtomKey} />
        <Filters 
          columnAtomKey={columnAtomKey}
          regions={formattedRegions}
          rowSelection={rowSelection}
          reservedTermOptions={[
            { value: "1yr", label: "1 year" },
            { value: "3yr", label: "3 years" }
          ]}
        />
        <div className="flex-1 min-h-0">
          <InstanceTable
            instances={allInstances}
            columnAtomKey={columnAtomKey}
            rowSelection={rowSelection}
            setRowSelection={setRowSelection}
            instanceCount={instanceCount}
          />
        </div>
      </div>
    </MarketingWrapper>
  );
}