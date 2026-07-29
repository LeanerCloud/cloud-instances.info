import { describe, expect, test } from "vitest";
import {
    commitmentTypeLabel,
    databaseSavingsPlanSupported,
    reservedTermOptions,
} from "./dataMappings";

describe("database savings plans", () => {
    test("databaseSavingsPlanSupported builds a database dropdown option", () => {
        const options = reservedTermOptions([...databaseSavingsPlanSupported]);
        const databaseOption = options.find(
            (o) => o.value === "yrTerm1DatabaseSavings.noUpfront",
        );
        expect(databaseOption).toEqual({
            value: "yrTerm1DatabaseSavings.noUpfront",
            label: "1-year Database Savings Plan - No Upfront",
            group: "Database Savings Plan",
        });
    });

    test("commitmentTypeLabel recognizes database savings term keys", () => {
        expect(commitmentTypeLabel("yrTerm1DatabaseSavings.noUpfront")).toBe(
            "Database Savings Plan",
        );
        expect(commitmentTypeLabel("DatabaseSavings.noUpfront")).toBe(
            "Database Savings Plan",
        );
        expect(commitmentTypeLabel("yrTerm1Savings.noUpfront")).toBe(
            "Compute Savings Plan",
        );
    });
});
