import { expect, test } from "vitest";
import { rdsEngineBucket, type RdsEnginePricing } from "./rdsPricing";

const postgresPricing: RdsEnginePricing = {
    ondemand: "0.178",
    reserved: { "yrTerm1Standard.noUpfront": "0.12" },
    multi_az: {
        ondemand: "0.356",
        reserved: { "yrTerm1Standard.noUpfront": "0.24" },
    },
};

const pricingByEngine: Record<string, RdsEnginePricing> = {
    "14": postgresPricing,
};

test("single-az returns root engine pricing", () => {
    const bucket = rdsEngineBucket(pricingByEngine, "14", "single-az");
    expect(bucket?.ondemand).toBe("0.178");
    expect(bucket?.reserved?.["yrTerm1Standard.noUpfront"]).toBe("0.12");
});

test("multi-az returns nested multi_az pricing", () => {
    const bucket = rdsEngineBucket(pricingByEngine, "14", "multi-az");
    expect(bucket?.ondemand).toBe("0.356");
    expect(bucket?.reserved?.["yrTerm1Standard.noUpfront"]).toBe("0.24");
});

test("missing engine key returns undefined", () => {
    expect(rdsEngineBucket(pricingByEngine, "2", "single-az")).toBeUndefined();
});

test("missing multi_az returns undefined in multi-az mode", () => {
    const singleOnly: Record<string, RdsEnginePricing> = {
        "14": { ondemand: "0.178" },
    };
    expect(rdsEngineBucket(singleOnly, "14", "multi-az")).toBeUndefined();
});

test("undefined region pricing returns undefined", () => {
    expect(rdsEngineBucket(undefined, "14", "single-az")).toBeUndefined();
});
