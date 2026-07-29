import { PlatformPricing, Pricing, RdsDeploymentOption } from "@/types";

// RDS scraper nests Multi-AZ ondemand/reserved under multi_az per engine. Kept
// here (not types.ts) so shared EC2/PlatformPricing isn't modified.
export type RdsEnginePricing = PlatformPricing & {
    multi_az?: {
        ondemand: string;
        reserved?: Record<string, string>;
    };
};

export function rdsEngineBucket(
    pricingByEngine: Pricing[string] | undefined,
    engineKey: string,
    deployment: RdsDeploymentOption,
): PlatformPricing | undefined {
    const engine = pricingByEngine?.[engineKey];
    if (!engine) {
        return undefined;
    }
    if (deployment === "multi-az") {
        return (engine as RdsEnginePricing).multi_az;
    }
    return engine;
}
