<template>
  <div>
    <div class="mb-6">
      <button @click="navigateTo('/conversations')" class="text-primary-400 hover:text-primary-300 mb-4">
        ← Tillbaka till konversationer
      </button>
      <h1 class="page-header">Konversation</h1>
      <p v-if="conversation" class="text-slate-400 mt-2">
        {{ conversation.listing_title }} - {{ conversation.marketplace_name }}
      </p>
    </div>

    <div v-if="loading" class="text-center py-12 text-slate-500">
      Laddar meddelanden...
    </div>

    <div v-else class="space-y-6">
      <!-- Message History -->
      <div class="card p-6">
        <h2 class="text-xl font-semibold text-slate-100 mb-4">Meddelandehistorik</h2>
        
        <div v-if="messages.length === 0" class="text-center py-8 text-slate-500">
          Inga meddelanden än
        </div>
        
        <div v-else class="space-y-4">
          <div 
            v-for="msg in messages" 
            :key="msg.id"
            :class="messageClass(msg)"
          >
            <div class="flex justify-between items-start mb-2">
              <span class="font-medium">
                {{ msg.direction === 'outgoing' ? 'Du' : 'Säljare' }}
              </span>
              <div class="flex items-center gap-2">
                <span :class="messageStatusClass(msg.status)">
                  {{ statusLabel(msg.status) }}
                </span>
                <span class="text-xs text-slate-500">
                  {{ formatDate(msg.created_at) }}
                </span>
              </div>
            </div>
            <p class="text-slate-200">{{ msg.content }}</p>
            
            <!-- Actions for pending messages -->
            <div v-if="msg.status === 'pending' && msg.direction === 'outgoing'" class="mt-4 flex gap-2">
              <button 
                @click="editMessage(msg)"
                class="btn btn-secondary text-sm"
              >
                Redigera
              </button>
              <button 
                @click="approveMessage(msg.id)"
                class="btn btn-primary text-sm"
              >
                Godkänn och skicka
              </button>
              <button 
                @click="rejectMessage(msg.id)"
                class="btn btn-danger text-sm"
              >
                Neka
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="flex gap-4">
        <button 
          @click="generateInitialMessage"
          v-if="messages.length === 0"
          class="btn btn-primary"
          :disabled="generatingMessage"
        >
          {{ generatingMessage ? 'Genererar...' : 'Generera första meddelande' }}
        </button>
        <button 
          @click="generateReply"
          v-else-if="canGenerateReply"
          class="btn btn-primary"
          :disabled="generatingMessage"
        >
          {{ generatingMessage ? 'Genererar...' : 'Generera svar' }}
        </button>
        <button 
          @click="showAddIncomingModal = true"
          class="btn btn-secondary"
        >
          Lägg till inkommande meddelande
        </button>
      </div>
    </div>

    <!-- Edit Message Modal -->
    <div v-if="editingMessage" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
      <div class="bg-slate-800 rounded-lg p-6 w-full max-w-2xl border border-slate-700">
        <h2 class="text-xl font-bold text-slate-100 mb-4">Redigera meddelande</h2>
        
        <form @submit.prevent="saveMessage">
          <div class="mb-4">
            <label class="label">Meddelande</label>
            <textarea 
              v-model="editForm.content"
              class="input h-32"
              required
            ></textarea>
          </div>
          
          <div class="flex justify-end gap-2">
            <button type="button" @click="editingMessage = null" class="btn btn-secondary">
              Avbryt
            </button>
            <button type="submit" class="btn btn-primary">
              Spara
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Add Incoming Message Modal -->
    <div v-if="showAddIncomingModal" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
      <div class="bg-slate-800 rounded-lg p-6 w-full max-w-2xl border border-slate-700">
        <h2 class="text-xl font-bold text-slate-100 mb-4">Lägg till inkommande meddelande</h2>
        
        <form @submit.prevent="addIncomingMessage">
          <div class="mb-4">
            <label class="label">Meddelande från säljare</label>
            <textarea 
              v-model="incomingForm.content"
              class="input h-32"
              required
            ></textarea>
          </div>
          
          <div class="flex justify-end gap-2">
            <button type="button" @click="showAddIncomingModal = false" class="btn btn-secondary">
              Avbryt
            </button>
            <button type="submit" class="btn btn-primary">
              Lägg till
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'

interface Message {
  id: number
  conversation_id: number
  direction: 'incoming' | 'outgoing'
  content: string
  status: string
  approved_at?: string
  sent_at?: string
  created_at: string
  updated_at: string
}

interface Conversation {
  id: number
  listing_id: number
  marketplace_id: number
  status: string
  created_at: string
  updated_at: string
  listing_title: string
  listing_price?: number
  marketplace_name: string
  pending_count: number
}

const route = useRoute()
const api = useApi()
const conversationId = computed(() => parseInt(route.params.id as string))

const conversation = ref<Conversation | null>(null)
const messages = ref<Message[]>([])
const loading = ref(true)
const generatingMessage = ref(false)
const editingMessage = ref<Message | null>(null)
const showAddIncomingModal = ref(false)

const editForm = ref({
  content: ''
})

const incomingForm = ref({
  content: ''
})

const canGenerateReply = computed(() => {
  const lastMessage = messages.value[messages.value.length - 1]
  return lastMessage?.direction === 'incoming' && !hasPendingOutgoing.value
})

const hasPendingOutgoing = computed(() => {
  return messages.value.some(m => m.direction === 'outgoing' && m.status === 'pending')
})

const fetchConversation = async () => {
  try {
    conversation.value = await api.get<Conversation>(`/conversations/${conversationId.value}`)
  } catch (error) {
    console.error('Failed to fetch conversation:', error)
  }
}

const fetchMessages = async () => {
  loading.value = true
  try {
    messages.value = await api.get<Message[]>(`/conversations/${conversationId.value}/messages`)
  } catch (error) {
    console.error('Failed to fetch messages:', error)
  } finally {
    loading.value = false
  }
}

const generateInitialMessage = async () => {
  if (!conversation.value) return
  
  generatingMessage.value = true
  try {
    await api.post('/messages', {
      listing_id: conversation.value.listing_id,
      message_type: 'initial'
    })
    await fetchMessages()
  } catch (error) {
    console.error('Failed to generate message:', error)
    alert('Kunde inte generera meddelande. Försök igen.')
  } finally {
    generatingMessage.value = false
  }
}

const generateReply = async () => {
  generatingMessage.value = true
  try {
    await api.post('/messages', {
      conversation_id: conversationId.value,
      message_type: 'reply'
    })
    await fetchMessages()
  } catch (error) {
    console.error('Failed to generate reply:', error)
    alert('Kunde inte generera svar. Försök igen.')
  } finally {
    generatingMessage.value = false
  }
}

const approveMessage = async (messageId: number) => {
  try {
    await api.put(`/messages/${messageId}/approve`, {})
    await fetchMessages()
  } catch (error) {
    console.error('Failed to approve message:', error)
    alert('Kunde inte godkänna meddelande. Försök igen.')
  }
}

const rejectMessage = async (messageId: number) => {
  if (!confirm('Är du säker på att du vill neka detta meddelande?')) return
  
  try {
    await api.put(`/messages/${messageId}/reject`, {})
    await fetchMessages()
  } catch (error) {
    console.error('Failed to reject message:', error)
    alert('Kunde inte neka meddelande. Försök igen.')
  }
}

const editMessage = (msg: Message) => {
  editingMessage.value = msg
  editForm.value.content = msg.content
}

const saveMessage = async () => {
  if (!editingMessage.value) return
  
  try {
    await api.put(`/messages/${editingMessage.value.id}`, {
      content: editForm.value.content
    })
    editingMessage.value = null
    await fetchMessages()
  } catch (error) {
    console.error('Failed to update message:', error)
    alert('Kunde inte uppdatera meddelande. Försök igen.')
  }
}

const addIncomingMessage = async () => {
  try {
    await api.post('/messages', {
      conversation_id: conversationId.value,
      message_type: 'incoming',
      content: incomingForm.value.content
    })
    incomingForm.value.content = ''
    showAddIncomingModal.value = false
    await fetchMessages()
  } catch (error) {
    console.error('Failed to add incoming message:', error)
    alert('Kunde inte lägga till meddelande. Försök igen.')
  }
}

const messageClass = (msg: Message) => {
  const baseClass = 'p-4 rounded-lg border'
  if (msg.direction === 'outgoing') {
    return `${baseClass} bg-primary-900/20 border-primary-700`
  } else {
    return `${baseClass} bg-slate-700/50 border-slate-600`
  }
}

const messageStatusClass = (status: string) => {
  const classes: Record<string, string> = {
    pending: 'px-2 py-1 bg-yellow-500/20 text-yellow-400 text-xs rounded',
    approved: 'px-2 py-1 bg-green-500/20 text-green-400 text-xs rounded',
    sent: 'px-2 py-1 bg-blue-500/20 text-blue-400 text-xs rounded',
    rejected: 'px-2 py-1 bg-red-500/20 text-red-400 text-xs rounded',
    received: 'px-2 py-1 bg-slate-500/20 text-slate-400 text-xs rounded'
  }
  return classes[status] || 'px-2 py-1 bg-slate-500/20 text-slate-400 text-xs rounded'
}

const statusLabel = (status: string) => {
  const labels: Record<string, string> = {
    pending: 'Väntar på granskning',
    approved: 'Godkänd',
    sent: 'Skickad',
    rejected: 'Nekad',
    received: 'Mottagen'
  }
  return labels[status] || status
}

const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  return new Intl.DateTimeFormat('sv-SE', { 
    year: 'numeric', 
    month: 'short', 
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

onMounted(async () => {
  await fetchConversation()
  await fetchMessages()
})
</script>

<style scoped>
.page-header {
  @apply text-3xl font-bold text-slate-100;
}

.card {
  @apply bg-slate-800 rounded-lg border border-slate-700;
}

.btn {
  @apply px-4 py-2 rounded-lg font-medium transition-colors;
}

.btn-primary {
  @apply bg-primary-600 text-white hover:bg-primary-700;
}

.btn-secondary {
  @apply bg-slate-700 text-slate-200 hover:bg-slate-600;
}

.btn-danger {
  @apply bg-red-600 text-white hover:bg-red-700;
}

.label {
  @apply block text-sm font-medium text-slate-300 mb-1;
}

.input {
  @apply w-full bg-slate-700 border border-slate-600 rounded-lg px-3 py-2 text-slate-100 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent;
}
</style>
