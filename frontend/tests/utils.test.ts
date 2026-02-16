import { describe, it, expect } from 'vitest'

describe('Utils', () => {
  describe('formatCurrency', () => {
    const formatCurrency = (cents: number) => `${(cents / 100).toFixed(2)} kr`

    it('should format öre to kr correctly', () => {
      expect(formatCurrency(10000)).toBe('100.00 kr')
      expect(formatCurrency(500)).toBe('5.00 kr')
      expect(formatCurrency(0)).toBe('0.00 kr')
    })

    it('should handle negative values', () => {
      expect(formatCurrency(-5000)).toBe('-50.00 kr')
    })
  })

  describe('formatDate', () => {
    const formatDate = (dateStr: string) => new Date(dateStr).toLocaleDateString('sv-SE')

    it('should format ISO date to Swedish format', () => {
      expect(formatDate('2024-01-15')).toBe('2024-01-15')
    })
  })

  describe('calculateProfit', () => {
    const calculateProfit = (item: any) => {
      const sellTotal = (item.sell_price || 0) + (item.sell_shipping_collected || 0)
      const buyTotal = item.buy_price + item.buy_shipping_cost
      const sellCost = (item.sell_packaging_cost || 0) + (item.sell_postage_cost || 0)
      return sellTotal - (buyTotal + sellCost)
    }

    it('should calculate profit correctly', () => {
      const item = {
        buy_price: 5000,
        buy_shipping_cost: 500,
        sell_price: 8000,
        sell_packaging_cost: 200,
        sell_postage_cost: 100,
        sell_shipping_collected: 0
      }
      // Sell: 8000, Buy: 5500, Costs: 300, Profit: 2200
      expect(calculateProfit(item)).toBe(2200)
    })

    it('should handle missing values', () => {
      const item = {
        buy_price: 5000,
        buy_shipping_cost: 0,
        sell_price: null,
        sell_packaging_cost: null,
        sell_postage_cost: null,
        sell_shipping_collected: null
      }
      expect(calculateProfit(item)).toBe(-5000)
    })
  })
})
