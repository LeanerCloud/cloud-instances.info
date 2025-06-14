"use client";

import { InstanceTable } from "@/components/InstanceTable";
import { Filters } from "@/components/Filters";
import { DoMigration } from "@/components/DoMigration";
import { MarketingWrapper } from "@/components/MarketingWrapper";
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

  return (
    <MarketingWrapper>
      <div className="flex flex-col h-[calc(100vh-4rem)] overflow-hidden">
        <DoMigration />
        <Filters instances={instances} />
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