import { PricingUnit } from "@/types";
import { ColumnDef } from "@tanstack/react-table";
import {
    doAllDataTablesMigrations,
    regex,
    makeSchemaWithDefaults,
    makeCellWithRegexSorter,
    expr,
} from "./shared";
import { CostDuration } from "@/types";
import RegionLinkPreloader from "@/components/RegionLinkPreloader";
import exprCompiler from "@/utils/expr";

export interface GCPPricing {
    [region: string]: {
        [platform: string]: {
            ondemand: number;
            preemptible?: number;
            spot?: number;
            committed?: {
                [term: string]: number;
            };
        };
    };
}

interface GCPStorage {
    local_ssd?: number;
    persistent_disk?: number;
    type?: string;
}

export type GCPInstance = {
    instance_type: string;
    pretty_name: string;
    pricing: GCPPricing;
    vcpu: number;
    memory: number;
    gpu?: {
        count: number;
        type: string;
    };
    family: string;
    arch: string[];
    storage?: GCPStorage;
    network_performance: string;
    sustained_use_discount: boolean;
    preemptible: boolean;
    sole_tenant?: boolean;
    custom_sizing?: boolean;
    machine_type?: string;
    generation?: string;
};

const initialColumnsArr = [
    ["pretty_name", true],
    ["instance_type", true],
    ["memory", true],
    ["vcpu", true],
    ["memory_per_vcpu", false],
    ["gpu", false],
    ["storage", true],
    ["network_performance", false],
    ["linux-ondemand", true],
    ["linux-committed", true],
    ["linux-preemptible", true],
    ["windows-ondemand", true],
    ["windows-committed", true],
] as const;

export const initialColumnsValue: {
    [idx in (typeof initialColumnsArr)[number][0]]: boolean;
} = {} as any;
for (const [key, value] of initialColumnsArr) {
    initialColumnsValue[key] = value;
}

export function makeColumnVisibilitySchema() {
    return makeSchemaWithDefaults(initialColumnsValue);
}

export function doDataTablesMigration() {
    return doAllDataTablesMigrations(
        "/gcp/",
        initialColumnsArr,
        initialColumnsValue,
    );
}

export function makePrettyNames<V>(
    makeColumnOption: (
        key: keyof typeof initialColumnsValue,
        label: string,
    ) => V,
) {
    return [
        makeColumnOption("pretty_name", "Name"),
        makeColumnOption("instance_type", "Machine Type"),
        makeColumnOption("memory", "Instance Memory"),
        makeColumnOption("vcpu", "vCPUs"),
        makeColumnOption("memory_per_vcpu", "Memory per vCPU"),
        makeColumnOption("gpu", "GPUs"),
        makeColumnOption("storage", "Storage"),
        makeColumnOption("network_performance", "Network"),
        makeColumnOption("linux-ondemand", "Linux On-Demand"),
        makeColumnOption("linux-committed", "Linux Committed"),
        makeColumnOption("linux-preemptible", "Linux Preemptible"),
        makeColumnOption("windows-ondemand", "Windows On-Demand"),
        makeColumnOption("windows-committed", "Windows Committed"),
    ];
}

function round(value: number) {
    return Math.round(value * 100) / 100;
}

function calculateCost(
    price: string | number | undefined,
    instance: GCPInstance,
    pricingUnit: PricingUnit,
    costDuration: CostDuration,
): number {
    if (!price) return -1;

    const hourMultipliers = {
        secondly: 1 / (60 * 60),
        minutely: 1 / 60,
        hourly: 1,
        daily: 24,
        weekly: 7 * 24,
        monthly: (365 * 24) / 12,
        annually: 365 * 24,
    };

    const durationMultiplier = hourMultipliers[costDuration];
    let pricingUnitModifier = 1;

    if (pricingUnit !== "instance") {
        pricingUnitModifier = instance[
            pricingUnit === "vcpu"
                ? "vcpu"
                : pricingUnit === "ecu"
                  ? "vcpu" // GCP doesn't have ECU, use vCPU
                  : "memory"
        ] as number;
    }

    return (Number(price) * durationMultiplier) / pricingUnitModifier;
}

export function calculateAndFormatCost(
    price: string | number | undefined,
    instance: GCPInstance,
    pricingUnit: PricingUnit,
    costDuration: CostDuration,
): string {
    const perTime = calculateCost(price, instance, pricingUnit, costDuration);
    if (perTime === -1) return "unavailable";

    const precision =
        costDuration === "secondly" || costDuration === "minutely" ? 6 : 4;

    const measuringUnits = {
        instances: "",
        vcpu: "vCPU",
        ecu: "vCPU",
        memory: "GiB",
    };

    let durationText: string = costDuration;
    if (costDuration === "secondly") durationText = "per sec";
    if (costDuration === "minutely") durationText = "per min";

    const pricingMeasuringUnits =
        pricingUnit === "instance"
            ? ` ${durationText}`
            : ` ${durationText} / ${measuringUnits[pricingUnit]}`;

    return `$${perTime.toFixed(precision)}${pricingMeasuringUnits}`;
}

function getPricingSorter(
    selectedRegion: string,
    pricingUnit: PricingUnit,
    costDuration: CostDuration,
    getter: (pricing: GCPPricing[string] | undefined) => number | undefined,
) {
    return {
        sortingFn: (rowA, rowB) => {
            const valueA = calculateCost(
                getter(rowA.original.pricing?.[selectedRegion]),
                rowA.original,
                pricingUnit,
                costDuration,
            );
            const valueB = calculateCost(
                getter(rowB.original.pricing?.[selectedRegion]),
                rowB.original,
                pricingUnit,
                costDuration,
            );
            return valueA - valueB;
        },
        sortUndefined: "last",
        accessorFn: (row) => {
            const g = getter(row.pricing?.[selectedRegion]);
            if (isNaN(Number(g))) return undefined;
            const value = calculateCost(g, row, pricingUnit, costDuration);
            return value === -1 ? undefined : value;
        },
        ...makeCellWithRegexSorter("pricing", (info) => {
            const pricing = info.row.original.pricing;
            const price = getter(pricing?.[selectedRegion]);
            if (isNaN(Number(price))) return "unavailable";
            return calculateAndFormatCost(
                price,
                info.row.original,
                pricingUnit,
                costDuration,
            );
        }),
    } satisfies Partial<ColumnDef<GCPInstance>>;
}

export const columnsGen = (
    selectedRegion: string,
    pricingUnit: PricingUnit,
    costDuration: CostDuration,
    committedUseTerm: string,
): ColumnDef<GCPInstance>[] => {
    return [
        {
            accessorKey: "pretty_name",
            id: "pretty_name",
            header: "Name",
            sortingFn: "alphanumeric",
            filterFn: regex({ accessorKey: "pretty_name" }),
        },
        {
            accessorKey: "instance_type",
            id: "instance_type",
            header: "Machine Type",
            sortingFn: "alphanumeric",
            cell: (info) => {
                const value = info.getValue() as string;
                return (
                    <RegionLinkPreloader
                        onClick={(e) => e.stopPropagation()}
                        href={`/gcp/compute/${value}`}
                    >
                        {value}
                    </RegionLinkPreloader>
                );
            },
            filterFn: regex({ accessorKey: "instance_type" }),
        },
        {
            accessorKey: "memory",
            header: "Instance Memory",
            size: 160,
            id: "memory",
            sortingFn: "alphanumeric",
            filterFn: expr,
            cell: (info) => `${info.getValue() as number} GiB`,
        },
        {
            accessorKey: "vcpu",
            header: "vCPUs",
            size: 160,
            id: "vcpu",
            sortingFn: "alphanumeric",
            filterFn: expr,
        },
        {
            accessorKey: "memory",
            header: "Memory per vCPU",
            size: 160,
            id: "memory_per_vcpu",
            filterFn: (row, _, filterValue) => {
                const value = row.original.memory;
                const cpu = row.original.vcpu;
                const memoryPerVcpu = round(value / cpu);
                try {
                    return exprCompiler(filterValue)(
                        memoryPerVcpu,
                        `${memoryPerVcpu} GiB/vCPU`,
                    );
                } catch {
                    return true;
                }
            },
            cell: (info) => {
                const value = info.getValue() as number;
                const cpu = info.row.original.vcpu;
                return `${round(value / cpu)} GiB/vCPU`;
            },
        },
        {
            accessorKey: "gpu",
            header: "GPUs",
            size: 160,
            id: "gpu",
            filterFn: expr,
            cell: (info) => {
                const gpu = info.getValue() as GCPInstance["gpu"];
                if (!gpu || gpu.count === 0) return "0";
                return `${gpu.count} × ${gpu.type}`;
            },
        },
        {
            accessorKey: "storage",
            header: "Storage",
            size: 160,
            id: "storage",
            filterFn: expr,
            cell: (info) => {
                const storage = info.getValue() as GCPStorage;
                if (!storage) return "No local storage";
                if (storage.local_ssd) return `${storage.local_ssd} GiB Local SSD`;
                return "Persistent disk only";
            },
        },
        {
            accessorKey: "network_performance",
            header: "Network Performance",
            size: 160,
            id: "network_performance",
            filterFn: regex({ accessorKey: "network_performance" }),
        },
        {
            accessorKey: "pricing",
            header: "Linux On Demand cost",
            id: "linux-ondemand",
            ...getPricingSorter(
                selectedRegion,
                pricingUnit,
                costDuration,
                (pricing) => pricing?.linux?.ondemand,
            ),
        },
        {
            accessorKey: "pricing",
            header: "Linux Committed Use",
            id: "linux-committed",
            ...getPricingSorter(
                selectedRegion,
                pricingUnit,
                costDuration,
                (pricing) => pricing?.linux?.committed?.[committedUseTerm],
            ),
        },
        {
            accessorKey: "pricing",
            header: "Linux Preemptible cost",
            id: "linux-preemptible",
            ...getPricingSorter(
                selectedRegion,
                pricingUnit,
                costDuration,
                (pricing) => pricing?.linux?.preemptible,
            ),
        },
        {
            accessorKey: "pricing",
            header: "Windows On Demand cost",
            id: "windows-ondemand",
            ...getPricingSorter(
                selectedRegion,
                pricingUnit,
                costDuration,
                (pricing) => pricing?.windows?.ondemand,
            ),
        },
        {
            accessorKey: "pricing",
            header: "Windows Committed Use",
            id: "windows-committed",
            ...getPricingSorter(
                selectedRegion,
                pricingUnit,
                costDuration,
                (pricing) => pricing?.windows?.committed?.[committedUseTerm],
            ),
        },
    ];
};