export const primeTablePt = {
  thead: {
    class: 'prism-table-head',
  },
  headerCell: {
    class: 'prism-table-header-cell',
  },
  columnHeaderContent: {
    class: 'prism-table-header-content',
  },
  bodyCell: {
    class: 'prism-table-body-cell',
  },
}

const currencyColumnTokens = [
  'amount',
  'alokasi',
  'biaya',
  'counterpart',
  'disbursement',
  'grant',
  'hibah',
  'idr',
  'jasa',
  'konstruksi',
  'loan',
  'lokal',
  'nilai',
  'pendanaan',
  'pinjaman',
  'realisasi',
  'usd',
]

export function isCurrencyColumn(field: string, header: string) {
  const searchable = `${field} ${header}`.toLowerCase()

  return currencyColumnTokens.some((token) => searchable.includes(token))
}

export function tableAlignClass(align: 'left' | 'center' | 'right') {
  return `prism-table-align-${align}`
}

export function tableCellClasses(options: {
  align: 'left' | 'center' | 'right'
  currency?: boolean
  nowrap?: boolean
}) {
  return [
    tableAlignClass(options.align),
    options.currency ? 'prism-table-currency' : '',
    options.nowrap ? 'prism-table-nowrap' : '',
  ]
    .filter(Boolean)
    .join(' ')
}
