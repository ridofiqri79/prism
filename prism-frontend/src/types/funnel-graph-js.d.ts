declare module 'funnel-graph-js/dist/js/funnel-graph.js' {
  export interface FunnelGraphData {
    labels?: string[]
    colors?: string | string[]
    values: number[] | number[][]
    subLabels?: string[]
  }

  export interface FunnelGraphOptions {
    container: string | HTMLElement
    data: FunnelGraphData
    direction?: 'horizontal' | 'vertical'
    gradientDirection?: 'horizontal' | 'vertical'
    displayPercent?: boolean
    width?: number
    height?: number
    subLabelValue?: 'percent' | 'raw'
  }

  export default class FunnelGraph {
    constructor(options: FunnelGraphOptions)
    draw(): void
    update(options: Partial<FunnelGraphOptions>): void
    updateData(data: FunnelGraphData): void
    setWidth(width: number): this
    setHeight(height: number): this
    makeVertical(): boolean
    makeHorizontal(): boolean
  }
}
