export interface Product {
  id: number
  brand: string | null
  name: string | null
  category: string | null
  model_variant: string | null
  sell_packaging_cost: number
  sell_postage_cost: number
  new_price?: number | null
  enabled?: boolean | null
  created_at?: string | null
}

export interface TradedItem {
  id: number
  product_id: number | null
  storage: number | null
  color_id: number | null
  buy_price: number
  buy_shipping_cost: number
  buy_transaction_id: number | null
  buy_date: string | null
  sell_price: number | null
  sell_packaging_cost: number | null
  sell_postage_cost: number | null
  sell_shipping_collected: number | null
  sell_transaction_id: number | null
  sell_date: string | null
  status_id: number
  source_link: string
  created_at: string
  listing_id: number | null
  product?: Product
}

export interface Listing {
  id: number
  product_id: number | null
  price: number | null
  valuation: number | null
  link: string
  condition_id: number | null
  shipping_cost: number | null
  title: string
  description: string
  marketplace_id: number | null
  status: string
  publication_date: string | null
  sold_date: string | null
  created_at: string
  is_my_listing: boolean
  eligible_for_shipping: boolean | null
  seller_pays_shipping: boolean | null
  buy_now: boolean | null
}

export interface Transaction {
  id: number
  date: string
  amount: number
  transaction_type: number | null
}

export interface TransactionType {
  id: number
  name: string
}

export interface Marketplace {
  id: number
  name: string
  link: string
}

export interface Condition {
  id: number
  title: string
}

export interface Color {
  id: number
  name: string
}

export interface TradeStatus {
  id: number
  name: string
}

export interface SearchTerm {
  id: number
  description: string
  url: string
  marketplace_id: number | null
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface SearchCriteria {
  id: number
  search_term_id: number
  marketplace_id: number | null
  max_price: number | null
  min_condition: number | null
  extra_params: string
  created_at: string
}

export interface SearchTermWithCriteria {
  search_term: SearchTerm
  criteria: SearchCriteria[]
}

export interface SearchHistory {
  id: number
  search_term_id: number
  search_term_desc: string
  url: string
  results_found: number
  new_ads_found: number
  marketplace_id: number | null
  marketplace_name: string
  searched_at: string
  created_at: string
}

export interface ScrapingRun {
  id: number
  started_at: string
  completed_at: string | null
  status: string
  total_ads_found: number
  total_listings_saved: number
  total_good_buys: number
  error_message: string | null
  created_at: string
}

export interface Valuation {
  id: number
  product_id: number | null
  valuation_type_id: number | null
  valuation_type?: string
  valuation: number
  metadata: any
  created_at: string
}

export interface ValuationType {
  id: number
  name: string
  enabled?: boolean | null
}

export interface ProductValuationTypeConfig {
  product_id: number
  valuation_type_id: number
  is_active: boolean
  weight: number
}

export interface ListingWithDetails {
  Listing: Listing
  Product: Product | null
  Valuations: Valuation[]
  PotentialProfit?: number
  DiscountPercent?: number
  ComputedValuation?: number
}

export const TRADE_STATUSES: Record<number, string> = {
  1: 'potential',
  2: 'purchased',
  3: 'in_stock',
  4: 'listed',
  5: 'sold'
}

export const LISTING_STATUSES = ['draft', 'active', 'sold', 'archived']
