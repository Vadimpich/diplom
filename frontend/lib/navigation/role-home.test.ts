import { describe, expect, it } from "vitest";
import { getRoleHome } from "@/lib/navigation/role-home";

describe("getRoleHome", () => {
  it("maps admin to /admin", () => {
    expect(getRoleHome("admin")).toBe("/admin");
  });

  it("maps operator to /operator", () => {
    expect(getRoleHome("operator")).toBe("/operator");
  });

  it("falls back to /operator for unknown or nullish role input", () => {
    expect(getRoleHome(undefined)).toBe("/operator");
    expect(getRoleHome(null)).toBe("/operator");
    expect(getRoleHome("unknown")).toBe("/operator");
  });
});
