import { describe, it, expect } from 'vitest'

// Pure computation logic extracted for testing (mirrors products.vue implementation)
function isTypeActiveForProduct(
  productId: number,
  typeId: number,
  configsByProduct: Record<number, { product_id: number; valuation_type_id: number; is_active: boolean }[]>
): boolean {
  const configs = configsByProduct[productId]
  if (!configs || configs.length === 0) return true
  const config = configs.find(c => c.valuation_type_id === typeId)
  if (!config) return true
  return config.is_active
}

function computeWeightedValuation(
  productId: number,
  enabledTypes: { id: number }[],
  valuationsByProduct: Record<number, { valuation_type_id: number | null; valuation: number }[]>,
  weights: Record<number, number>,
  configsByProduct: Record<number, { product_id: number; valuation_type_id: number; is_active: boolean }[]> = {}
): { average: number; safetyPercent: number } | null {
  const activeTypes = enabledTypes.filter(vt => isTypeActiveForProduct(productId, vt.id, configsByProduct))
  if (activeTypes.length === 0) return null
  const vals = valuationsByProduct[productId] ?? []
  const getVal = (typeId: number) => vals.find(v => v.valuation_type_id === typeId) ?? null

  const entries = activeTypes
    .map(vt => {
      const v = getVal(vt.id)
      return v !== null ? { valuation: v.valuation, weight: weights[vt.id] ?? 1 } : null
    })
    .filter((e): e is { valuation: number; weight: number } => e !== null)

  if (entries.length === 0) return null
  const totalWeight = entries.reduce((s, e) => s + e.weight, 0)
  if (totalWeight === 0) return null
  const average = entries.reduce((s, e) => s + e.valuation * e.weight, 0) / totalWeight
  let safetyPercent = 100
  if (entries.length > 1) {
    const mean = entries.reduce((s, e) => s + e.valuation, 0) / entries.length
    const variance = entries.reduce((s, e) => s + Math.pow(e.valuation - mean, 2), 0) / entries.length
    const stdDev = Math.sqrt(variance)
    safetyPercent = mean !== 0 ? Math.max(0, Math.round(100 - (stdDev / Math.abs(mean) * 100))) : 0
  }
  return { average: Math.round(average), safetyPercent }
}

describe('isTypeActiveForProduct', () => {
  it('returns true when no configs exist for product (backward compatible)', () => {
    expect(isTypeActiveForProduct(1, 1, {})).toBe(true)
  })

  it('returns true when configs are empty for product', () => {
    expect(isTypeActiveForProduct(1, 1, { 1: [] })).toBe(true)
  })

  it('returns true when no config for specific type', () => {
    const configs = { 1: [{ product_id: 1, valuation_type_id: 2, is_active: false }] }
    expect(isTypeActiveForProduct(1, 1, configs)).toBe(true)
  })

  it('returns false when type is deactivated for product', () => {
    const configs = { 1: [{ product_id: 1, valuation_type_id: 1, is_active: false }] }
    expect(isTypeActiveForProduct(1, 1, configs)).toBe(false)
  })

  it('returns true when type is explicitly activated for product', () => {
    const configs = { 1: [{ product_id: 1, valuation_type_id: 1, is_active: true }] }
    expect(isTypeActiveForProduct(1, 1, configs)).toBe(true)
  })
})

describe('computeWeightedValuation', () => {
  it('returns null when no enabled valuation types', () => {
    expect(computeWeightedValuation(1, [], {}, {})).toBeNull()
  })

  it('returns null when product has no valuations', () => {
    const types = [{ id: 1 }, { id: 2 }]
    const result = computeWeightedValuation(1, types, {}, { 1: 1, 2: 1 })
    expect(result).toBeNull()
  })

  it('returns correct average with equal weights', () => {
    const types = [{ id: 1 }, { id: 2 }]
    const vbp = { 1: [{ valuation_type_id: 1, valuation: 1000 }, { valuation_type_id: 2, valuation: 2000 }] }
    const result = computeWeightedValuation(1, types, vbp, { 1: 1, 2: 1 })
    expect(result).not.toBeNull()
    expect(result!.average).toBe(1500)
  })

  it('returns correct average with custom weights', () => {
    const types = [{ id: 1 }, { id: 2 }]
    const vbp = { 1: [{ valuation_type_id: 1, valuation: 1000 }, { valuation_type_id: 2, valuation: 3000 }] }
    // weight 1 for type 1, weight 3 for type 2 => (1000*1 + 3000*3) / (1+3) = 10000/4 = 2500
    const result = computeWeightedValuation(1, types, vbp, { 1: 1, 2: 3 })
    expect(result!.average).toBe(2500)
  })

  it('returns 100% safety when only one valuation exists', () => {
    const types = [{ id: 1 }, { id: 2 }]
    const vbp = { 1: [{ valuation_type_id: 1, valuation: 1000 }] }
    const result = computeWeightedValuation(1, types, vbp, { 1: 1, 2: 1 })
    expect(result!.safetyPercent).toBe(100)
  })

  it('calculates lower safety when values diverge significantly', () => {
    const types = [{ id: 1 }, { id: 2 }]
    const vbp = { 1: [{ valuation_type_id: 1, valuation: 100 }, { valuation_type_id: 2, valuation: 900 }] }
    const result = computeWeightedValuation(1, types, vbp, { 1: 1, 2: 1 })
    // Large spread should result in lower safety
    expect(result!.safetyPercent).toBeLessThan(100)
  })

  it('returns null when total weights are zero', () => {
    const types = [{ id: 1 }]
    const vbp = { 1: [{ valuation_type_id: 1, valuation: 1000 }] }
    const result = computeWeightedValuation(1, types, vbp, { 1: 0 })
    expect(result).toBeNull()
  })

  it('ignores types without a valuation for the product', () => {
    const types = [{ id: 1 }, { id: 2 }]
    // Only type 1 has a valuation for product 1
    const vbp = { 1: [{ valuation_type_id: 1, valuation: 2000 }] }
    const result = computeWeightedValuation(1, types, vbp, { 1: 1, 2: 1 })
    expect(result!.average).toBe(2000)
    expect(result!.safetyPercent).toBe(100)
  })

  it('excludes deactivated types from weighted average', () => {
    const types = [{ id: 1 }, { id: 2 }]
    const vbp = { 1: [{ valuation_type_id: 1, valuation: 1000 }, { valuation_type_id: 2, valuation: 3000 }] }
    // Deactivate type 2 for product 1 => only type 1 used => average = 1000
    const configs = { 1: [{ product_id: 1, valuation_type_id: 2, is_active: false }] }
    const result = computeWeightedValuation(1, types, vbp, { 1: 1, 2: 1 }, configs)
    expect(result!.average).toBe(1000)
    expect(result!.safetyPercent).toBe(100)
  })

  it('returns null when all types deactivated for product', () => {
    const types = [{ id: 1 }, { id: 2 }]
    const vbp = { 1: [{ valuation_type_id: 1, valuation: 1000 }, { valuation_type_id: 2, valuation: 2000 }] }
    const configs = {
      1: [
        { product_id: 1, valuation_type_id: 1, is_active: false },
        { product_id: 1, valuation_type_id: 2, is_active: false }
      ]
    }
    const result = computeWeightedValuation(1, types, vbp, { 1: 1, 2: 1 }, configs)
    expect(result).toBeNull()
  })

  it('uses all types when no config (backward compatible)', () => {
    const types = [{ id: 1 }, { id: 2 }]
    const vbp = { 1: [{ valuation_type_id: 1, valuation: 1000 }, { valuation_type_id: 2, valuation: 2000 }] }
    const result = computeWeightedValuation(1, types, vbp, { 1: 1, 2: 1 })
    expect(result!.average).toBe(1500)
  })

  it('when last type is deactivated, remaining type gets full weight', () => {
    const types = [{ id: 1 }, { id: 2 }, { id: 3 }]
    const vbp = {
      1: [
        { valuation_type_id: 1, valuation: 1000 },
        { valuation_type_id: 2, valuation: 2000 },
        { valuation_type_id: 3, valuation: 3000 }
      ]
    }
    // Only type 3 active
    const configs = {
      1: [
        { product_id: 1, valuation_type_id: 1, is_active: false },
        { product_id: 1, valuation_type_id: 2, is_active: false },
        { product_id: 1, valuation_type_id: 3, is_active: true }
      ]
    }
    const result = computeWeightedValuation(1, types, vbp, { 1: 1, 2: 1, 3: 1 }, configs)
    expect(result!.average).toBe(3000)
  })
})
