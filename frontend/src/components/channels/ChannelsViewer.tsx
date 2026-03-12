import { useState } from 'react'
import { useChannels } from '../../hooks/useChannels'
import { Button } from '../ui/Button'
import { ChannelForm } from './ChannelForm'
import { ChannelDetailsModal } from './ChannelDetailsModal'
import { ChannelsTable } from './ChannelsTable'

export const ChannelsViewer: React.FC = () => {
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [selectedChannel, setSelectedChannel] = useState<any>(null)
  const [editingChannel, setEditingChannel] = useState<any>(null)

  const {
    data,
    loading,
    error,
    createChannel,
    updateChannel,
    deleteChannel,
    testChannel,
  } = useChannels()

  const channels = data?.channels || []

  const handleCreateChannel = async (formData: any) => {
    await createChannel(formData)
    setShowCreateForm(false)
  }

  const handleUpdateChannel = async (formData: any) => {
    if (editingChannel) {
      await updateChannel(editingChannel.id, formData)
      setEditingChannel(null)
    }
  }

  const handleDeleteChannel = async () => {
    if (selectedChannel) {
      await deleteChannel(selectedChannel.id)
      setSelectedChannel(null)
    }
  }

  const handleTestChannel = async () => {
    if (selectedChannel) {
      await testChannel(selectedChannel.id)
    }
  }

  if (error) {
    return (
      <div className="rounded-lg border border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-900/20 p-4">
        <div className="text-red-800 dark:text-red-200">Error: {error}</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-end">
        <Button variant="primary" onClick={() => setShowCreateForm(true)}>
          + Create Channel
        </Button>
      </div>

      <ChannelsTable
        channels={channels}
        loading={loading}
        onView={(channel) => setSelectedChannel(channel)}
      />

      {showCreateForm && (
        <ChannelForm
          isOpen={showCreateForm}
          onClose={() => setShowCreateForm(false)}
          onSubmit={handleCreateChannel}
        />
      )}

      {editingChannel && (
        <ChannelForm
          channel={editingChannel}
          isOpen={!!editingChannel}
          onClose={() => setEditingChannel(null)}
          onSubmit={handleUpdateChannel}
        />
      )}

      {selectedChannel && (
        <ChannelDetailsModal
          channel={selectedChannel}
          isOpen={!!selectedChannel}
          onClose={() => setSelectedChannel(null)}
          onEdit={() => {
            setEditingChannel(selectedChannel)
            setSelectedChannel(null)
          }}
          onDelete={handleDeleteChannel}
          onTest={handleTestChannel}
        />
      )}
    </div>
  )
}
