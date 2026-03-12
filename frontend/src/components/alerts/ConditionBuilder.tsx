import { Button } from '../ui/Button'
import { Input } from '../ui/Input'

interface Condition {
  id: string
  field: string
  operator: string
  value: string
}

interface ConditionBuilderProps {
  conditions: Condition[]
  onConditionsChange: (conditions: Condition[]) => void
}

const OPERATORS = ['equals', 'contains', 'greater_than', 'less_than', 'matches']
const FIELDS = ['log_level', 'log_message', 'error_code', 'response_time']

export const ConditionBuilder: React.FC<ConditionBuilderProps> = ({
  conditions,
  onConditionsChange,
}) => {
  const addCondition = () => {
    onConditionsChange([
      ...conditions,
      {
        id: Date.now().toString(),
        field: FIELDS[0],
        operator: OPERATORS[0],
        value: '',
      },
    ])
  }

  const updateCondition = (id: string, updates: Partial<Condition>) => {
    onConditionsChange(
      conditions.map((c) => (c.id === id ? { ...c, ...updates } : c))
    )
  }

  const removeCondition = (id: string) => {
    onConditionsChange(conditions.filter((c) => c.id !== id))
  }

  return (
    <div className="space-y-4">
      <label className="block text-sm font-medium text-slate-700 dark:text-slate-300">
        Alert Conditions
      </label>

      {conditions.map((condition, index) => (
        <div key={condition.id} className="flex gap-2 items-start">
          {index > 0 && (
            <div className="mt-2 px-2 text-sm font-medium text-slate-600 dark:text-slate-400">
              AND
            </div>
          )}

          <div className="flex-1 grid grid-cols-3 gap-2">
            <select
              value={condition.field}
              onChange={(e) => updateCondition(condition.id, { field: e.target.value })}
              className="px-3 py-2 border border-slate-300 rounded-lg dark:bg-slate-800 dark:border-slate-600 dark:text-white"
            >
              {FIELDS.map((field) => (
                <option key={field} value={field}>
                  {field.replace('_', ' ')}
                </option>
              ))}
            </select>

            <select
              value={condition.operator}
              onChange={(e) => updateCondition(condition.id, { operator: e.target.value })}
              className="px-3 py-2 border border-slate-300 rounded-lg dark:bg-slate-800 dark:border-slate-600 dark:text-white"
            >
              {OPERATORS.map((op) => (
                <option key={op} value={op}>
                  {op.replace('_', ' ')}
                </option>
              ))}
            </select>

            <Input
              placeholder="Value"
              value={condition.value}
              onChange={(e) => updateCondition(condition.id, { value: e.target.value })}
            />
          </div>

          <button
            onClick={() => removeCondition(condition.id)}
            className="px-3 py-2 text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
          >
            Remove
          </button>
        </div>
      ))}

      <Button variant="secondary" size="sm" onClick={addCondition}>
        + Add Condition
      </Button>
    </div>
  )
}
