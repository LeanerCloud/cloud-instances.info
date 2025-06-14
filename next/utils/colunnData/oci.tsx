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

export interface OCIPricing {
    [region: string]: {
        [platform: string]: {
            ondemand: number;
            reserved?: {
                [term: string]: number;
            };
        };
    };
}

interface OCIStorage {
    boot_volume?: number;
    block_volume?: number;
    local_storage?: number;
    type?: string;
}

export type OCIInstance = {
    instance_type: string;
    pretty_name: string;
    pricing: OCIPricing;
    vcpu: number;
    memory: number;
    gpu?: {
        count: number;
        type: string;
    };
    family: string;
    arch: string[];
    storage?: OCIStorage;
    network_performance: string;
    shape_type: string; // VM.Standard, BM.Standard, etc.
    processor?: string;
    baseline_ocpu_utilization?: string;
    max_bandwidth?: string;
    availability_domain?: string;
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
    ["shape_type", false],
    ["processor", false],
    ["linux-ondemand", true],
    ["linux-reserved", true],
    ["windows-ondemand", true],
    ["windows-reserved", true],
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
        "/oci/",
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
        makeColumnOption("instance_type", "Shape"),
        makeColumnOption("memory", "Instance Memory"),
        makeColumnOption("vcpu", "OCPUs"),
        makeColumnOption("memory_per_vcpu", "Memory per OCPU"),
        makeColumnOption("gpu", "GPUs"),
        makeColumnOption("storage", "Storage"),
        makeColumnOption("network_performance", "Network"),
        makeColumnOption("shape_type", "Shape Type"),
        makeColumnOption("processor", "Processor"),
        makeColumnOption("linux-ondemand", "Linux On-Demand"),
        makeColumnOption("linux-reserved", "Linux Reserved"),
        makeColumnOption("windows-ondemand", "Windows On-Demand"),
        makeColumnOption("windows-reserved", "Windows Reserved"),
    ];
}

function round(value: number) {
    return Math.round(value * 100) / 100;
}

function calculateCost(
    price: string | number | undefined,
    instance: OCIInstance,
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
                  ? "vcpu" // OCI uses OCPUs, similar to vCPUs
                  : "memory"
        ] as number;
    }

    return (Number(price) * durationMultiplier) / pricingUnitModifier;
}

export function calculateAndFormatCost(
    price: string | number | undefined,
    instance: OCIInstance,
    pricingUnit: PricingUnit,
    costDuration: CostDuration,
): string {
    const perTime = calculateCost(price, instance, pricingUnit, costDuration);
    if (perTime === -1) return "unavailable";

    const precision =
        costDuration === "secondly" || costDuration === "minutely" ? 6 : 4;

    const measuringUnits = {
        instances: "",
        vcpu: "OCPU",
        ecu: "OCPU",
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
    getter: (pricing: OCIPricing[string] | undefined) => number | undefined,
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
    } satisfies Partial<ColumnDef<OCIInstance>>;
}

export const columnsGen = (
    selectedRegion: string,
    pricingUnit: PricingUnit,
    costDuration: CostDuration,
    reservedTerm: string,
): ColumnDef<OCIInstance>[] => {
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
            header: "Shape",
            sortingFn: "alphanumeric",
            cell: (info) => {
                const value = info.getValue() as string;
                return (
                    <RegionLinkPreloader
                        onClick={(e) => e.stopPropagation()}
                        href={`/oci/compute/${value}`}
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
            header: "OCPUs",
            size: 160,
            id: "vcpu",
            sortingFn: "alphanumeric",
            filterFn: expr,
        },
        {
            accessorKey: "memory",
            header: "Memory per OCPU",
            size: 160,
            id: "memory_per_vcpu",
            filterFn: (row, _, filterValue) => {
                const value = row.original.memory;
                const cpu = row.original.vcpu;
                const memoryPerVcpu = round(value / cpu);
                try {
                    return exprCompiler(filterValue)(
                        memoryPerVcpu,
                        `${memoryPerVcpu} GiB/OCPU`,
                    );
                } catch {
                    return true;
                }
            },
            cell: (info) => {
                const value = info.getValue() as number;
                const cpu = info.row.original.vcpu;
                return `${round(value / cpu)} GiB/OCPU`;
            },
        },
        {
            accessorKey: "gpu",
            header: "GPUs",
            size: 160,
            id: "gpu",
            filterFn: expr,
            cell: (info) => {
                const gpu = info.getValue() as OCIInstance["gpu"];
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
                const storage = info.getValue() as OCIStorage;
                if (!storage) return "Block volume only";
                if (storage.local_storage) return `${storage.local_storage} GiB Local`;
                return "Block volume only";
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
            accessorKey: "shape_type",
            header: "Shape Type",
            size: 160,
            id: "shape_type",
            filterFn: regex({ accessorKey: "shape_type" }),
        },
        {
            accessorKey: "processor",
            header: "Processor",
            size: 160,
            id: "processor",
            filterFn: regex({ accessorKey: "processor" }),
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
            header: "Linux Reserved cost",
            id: "linux-reserved",
            ...getPricingSorter(
                selectedRegion,
                pricingUnit,
                costDuration,
                (pricing) => pricing?.linux?.reserved?.[reservedTerm],
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
            header: "Windows Reserved cost",
            id: "windows-reserved",
            ...getPricingSorter(
                selectedRegion,
                pricingUnit,
                costDuration,
                (pricing) => pricing?.windows?.reserved?.[reservedTerm],
            ),
        },
    ];
};