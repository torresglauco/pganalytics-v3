import { Button } from '../ui/Button'

interface MetricsControlsProps {
  timeRange: string
  onTimeRangeChange: (range: string) => void
}

export const MetricsControls: React.FC<MetricsControlsProps> = ({
  timeRange,
  onTimeRangeChange,
}) => {
  const ranges = ['24h', '7d', '30d']

  return (
    <div className="flex gap-2 flex-wrap">
      {ranges.map((range) => (
        <Button
          key={range}
          variant={timeRange === range ? 'primary' : 'secondary'}
          size="sm"
          onClick={() => onTimeRangeChange(range)}
        >
          {range === '24h' ? 'Last 24 Hours' : range === '7d' ? 'Last 7 Days' : 'Last 30 Days'}
        </Button>
      ))}
    </div>
  )
}
