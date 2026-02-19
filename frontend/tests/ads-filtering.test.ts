import { describe, it, expect } from 'vitest'

// Types matching the frontend
interface Listing {
  id: number
  price: number | null
  valuation: number | null
}

interface ListingWithDetails {
  Listing: Listing
  Valuations: { id?: number; valuation_type_id: number; valuation: number }[]
  PotentialProfit?: number
  DiscountPercent?: number
}

// Filter function - now just returns all (potential filtering done server-side)
function filterListingsByTab(
  listings: ListingWithDetails[], 
  tab: 'all' | 'potential'
): ListingWithDetails[] {
  if (tab === 'potential') {
    return listings.filter(l => 
      l.PotentialProfit !== undefined && 
      l.PotentialProfit > 0 &&
      l.DiscountPercent !== undefined &&
      l.DiscountPercent > 0
    )
  }
  return listings;
}

describe('Ads Page Filtering - Tab: Alla', () => {
  it('should return all listings when tab is "all"', () => {
    const listings: ListingWithDetails[] = [
      {
        Listing: { id: 1, price: 1000, valuation: 800 },
        Valuations: [{ valuation_type_id: 1, valuation: 800 }]
      },
      {
        Listing: { id: 2, price: 500, valuation: 1000 },
        Valuations: [{ valuation_type_id: 1, valuation: 1000 }]
      }
    ]
    
    const result = filterListingsByTab(listings, 'all')
    expect(result).toHaveLength(2)
  })

  it('should return empty array when no listings exist', () => {
    const listings: ListingWithDetails[] = []
    
    const resultAll = filterListingsByTab(listings, 'all')
    expect(resultAll).toHaveLength(0)
  })
})

describe('Ads Page Filtering - Tab: Potentiella', () => {
  it('should return only listings with positive potential profit', () => {
    const listings: ListingWithDetails[] = [
      {
        Listing: { id: 1, price: 1000, valuation: 800 },
        Valuations: [],
        PotentialProfit: 200,
        DiscountPercent: 20
      },
      {
        Listing: { id: 2, price: 500, valuation: 1000 },
        Valuations: [],
        PotentialProfit: -100,
        DiscountPercent: -10
      }
    ]
    
    const result = filterListingsByTab(listings, 'potential')
    expect(result).toHaveLength(1)
    expect(result[0].Listing.id).toBe(1)
  })

  it('should return empty when no listings have potential profit', () => {
    const listings: ListingWithDetails[] = [
      {
        Listing: { id: 1, price: 1000, valuation: 800 },
        Valuations: []
      }
    ]
    
    const result = filterListingsByTab(listings, 'potential')
    expect(result).toHaveLength(0)
  })
})
