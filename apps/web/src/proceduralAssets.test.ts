import { describe, expect, it } from "vitest";
import {
  proceduralPropFamily,
  propColor,
  tilePalette,
  wallColor,
} from "./proceduralAssets";
import type { SceneEntity } from "./types";

describe("procedural asset catalog", () => {
  it("classifies the built-in prop families deterministically", () => {
    expect(proceduralPropFamily(entity("prop-column"))).toBe("column");
    expect(proceduralPropFamily(entity("prop-brazier"))).toBe("brazier");
    expect(proceduralPropFamily(entity("prop-crate"))).toBe("crate");
    expect(proceduralPropFamily(entity("prop-chest", "key"))).toBe("chest");
    expect(proceduralPropFamily(entity("unknown"))).toBe("generic");
  });

  it("gives hazardous surfaces and secret edges distinct palettes", () => {
    expect(tilePalette("water")).not.toEqual(tilePalette("lava"));
    expect(tilePalette("lava").emissiveIntensity).toBeGreaterThan(1);
    expect(wallColor("secret-door")).not.toBe(wallColor("wall"));
    expect(propColor("chest", entity("prop-chest", "key"))).toBe("#d8b45e");
  });
});

function entity(assetId: string, kind: SceneEntity["kind"] = "prop"): SceneEntity {
  return {
    id: assetId,
    kind,
    name: { "pt-BR": assetId, "en-US": assetId },
    position: { x: 0, z: 0 },
    assetId,
    blocksMovement: false,
    hidden: false,
  };
}
