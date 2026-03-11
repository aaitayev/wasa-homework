<script setup>
import { ref, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import api from '../services/axios.js';
import ErrorMsg from '../components/ErrorMsg.vue';
import LoadingSpinner from '../components/LoadingSpinner.vue';

const props = defineProps(['id']);
const router = useRouter();

const messages = ref([]);
const isLoading = ref(false);
const errorMsg = ref('');
const inputText = ref('');
const replyingTo = ref(null);

const myUsername = localStorage.getItem('username') || '';

// Forwarding state
const showForwardModal = ref(false);
const forwardingMessageId = ref(null);
const availableConversations = ref([]);

async function loadConversation() {
  if (!props.id) return;
  
  errorMsg.value = '';
  isLoading.value = true;
  
  try {
    const res = await api.get(`/conversations/${props.id}`);
    const convData = res.data;
    if (convData && convData.messages) {
      messages.value = convData.messages.map(m => ({
        ...m,
        createdAtDate: new Date(m.createdAt)
      })).sort((a, b) => b.createdAtDate - a.createdAtDate);
    } else {
      messages.value = [];
    }
  } catch (error) {
    if (error.response && error.response.status === 401) {
      router.push('/login');
    } else {
      errorMsg.value = 'Failed to load conversation details.';
      console.error(error);
    }
  } finally {
    isLoading.value = false;
  }
}

watch(() => props.id, () => {
  loadConversation();
});

onMounted(() => {
  loadConversation();
});

function formatTime(dateObj) {
  if (!dateObj || dateObj.getTime() === 0) return '';
  return dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) + ' ' + dateObj.toLocaleDateString();
}

async function doSend() {
  const text = inputText.value.trim();
  if (!text) return;
  
  let outgoingText = text;
  if (replyingTo.value) {
    outgoingText = `↩︎ ${replyingTo.value.text.substring(0, 30)}...\n${text}`;
  }

  try {
    isLoading.value = true;
    await api.post('/messages', {
      conversationId: props.id,
      text: outgoingText,
      isGroup: false 
    });
    
    inputText.value = '';
    replyingTo.value = null;
    await loadConversation();
  } catch (error) {
    if (error.response && error.response.status === 401) {
      router.push('/login');
    } else {
      errorMsg.value = 'Failed to send message.';
    }
  } finally {
    isLoading.value = false;
  }
}

async function doDelete(msgId) {
  if (!confirm("Are you sure you want to delete this message?")) return;
  try {
    isLoading.value = true;
    await api.delete(`/messages/${msgId}`);
    await loadConversation();
  } catch (error) {
    errorMsg.value = "Failed to delete message";
  } finally {
    isLoading.value = false;
  }
}

async function doComment(msgId, emoji) {
  try {
    isLoading.value = true;
    await api.post(`/messages/${msgId}/comment`, { comment: emoji });
    await loadConversation();
  } catch (error) {
    errorMsg.value = "Failed to add reaction.";
  } finally {
    isLoading.value = false;
  }
}

async function doUncomment(msgId) {
  try {
    isLoading.value = true;
    await api.delete(`/messages/${msgId}/comment`);
    await loadConversation();
  } catch (error) {
    errorMsg.value = "Failed to remove reaction.";
  } finally {
    isLoading.value = false;
  }
}

async function openForwardModal(msgId) {
  forwardingMessageId.value = msgId;
  errorMsg.value = '';
  try {
    isLoading.value = true;
    const res = await api.get('/conversations');
    availableConversations.value = res.data.filter(c => c.id !== props.id);
    showForwardModal.value = true;
  } catch (error) {
    errorMsg.value = 'Failed to load conversations for forwarding.';
  } finally {
    isLoading.value = false;
  }
}

async function confirmForward(targetId) {
  try {
    isLoading.value = true;
    await api.post(`/messages/${forwardingMessageId.value}/forward`, { conversationId: targetId });
    showForwardModal.value = false;
    alert("Message forwarded!");
  } catch (error) {
    errorMsg.value = "Failed to forward message.";
  } finally {
    isLoading.value = false;
  }
}

function handleKeydown(e) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    doSend();
  }
}

const commonEmojis = ['👍', '❤️', '😂', '😮', '😢', '🔥'];
</script>

<template>
  <div class="conversation-container d-flex flex-column h-100 p-2 position-relative">
    <div class="header d-flex justify-content-between align-items-center mb-2 border-bottom pb-2">
      <div class="d-flex align-items-center">
        <button class="btn btn-link p-0 me-3 text-decoration-none" @click="router.push('/conversations')" title="Back">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" class="bi bi-arrow-left" viewBox="0 0 16 16">
            <path fill-rule="evenodd" d="M15 8a.5.5 0 0 0-.5-.5H2.707l3.147-3.146a.5.5 0 1 0-.708-.708l-4 4a.5.5 0 0 0 0 .708l4 4a.5.5 0 0 0 .708-.708L2.707 8.5H14.5A.5.5 0 0 0 15 8z"/>
          </svg>
        </button>
        <h4 class="mb-0 text-truncate" style="max-width: 250px;">Conv: {{ id }}</h4>
      </div>
      <div class="actions">
        <button class="btn btn-outline-secondary btn-sm me-2" @click="loadConversation" :disabled="isLoading">
          <svg v-if="!isLoading" xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-arrow-clockwise" viewBox="0 0 16 16">
            <path fill-rule="evenodd" d="M8 3a5 5 0 1 0 4.546 2.914.5.5 0 0 1 .908-.417A6 6 0 1 1 8 2v1z"/>
            <path d="M8 4.466V.534a.25.25 0 0 1 .41-.192l2.36 1.966c.12.1.12.284 0 .384L8.41 4.658A.25.25 0 0 1 8 4.466z"/>
          </svg>
          <span v-else class="spinner-border spinner-border-sm" role="status"></span>
        </button>
      </div>
    </div>
    
    <ErrorMsg v-if="errorMsg" :msg="errorMsg" class="mb-2" />
    
    <div class="messages-list flex-grow-1 overflow-auto d-flex flex-column-reverse mb-3 p-3 bg-white rounded shadow-inner" style="min-height: 200px;">
      <div v-if="messages.length === 0 && !isLoading" class="text-center text-muted my-auto">
        <div class="mb-2">No messages here yet.</div>
        <small>Be the first to say hello!</small>
      </div>

      <LoadingSpinner v-if="isLoading && messages.length === 0" :loading="true" />
      
      <div 
        v-for="msg in messages" 
        :key="msg.id"
        class="d-flex mb-3"
        :class="msg.senderId === myUsername ? 'justify-content-end' : 'justify-content-start'"
      >
        <div 
          class="message-bubble position-relative"
          :class="[
            msg.senderId === myUsername ? 'bg-primary text-white bubble-right' : 'bg-light text-dark bubble-left',
            msg.deleted ? 'opacity-75' : ''
          ]"
          style="max-width: 80%; padding: 10px 15px; border-radius: 18px;"
        >
          <div class="bubble-header d-flex justify-content-between align-items-center mb-1" style="font-size: 0.75rem;">
            <span class="fw-bold me-2">{{ msg.senderId === myUsername ? 'You' : msg.senderId }}</span>
            <span :class="msg.senderId === myUsername ? 'text-white-50' : 'text-muted'">{{ formatTime(msg.createdAtDate) }}</span>
          </div>

          <div v-if="msg.forwardedFrom" class="forwarded-tag mb-1" style="font-size: 0.65rem; font-style: italic; opacity: 0.8;">
            Forwarded from {{ msg.forwardedFrom }}
          </div>

          <div class="message-content">
            <span v-if="msg.deleted" class="fst-italic opacity-50">This message was deleted</span>
            <span v-else style="white-space: pre-wrap; word-break: break-word;">{{ msg.text }}</span>
          </div>

          <!-- Reactions -->
          <div v-if="msg.comment" class="reactions-container mt-2">
            <span 
              class="badge rounded-pill bg-white text-dark shadow-sm d-inline-flex align-items-center px-2 py-1"
              style="font-size: 0.9rem; cursor: pointer;"
              @click="doUncomment(msg.id)"
              title="Remove reaction"
            >
              {{ msg.comment }}
            </span>
          </div>

          <!-- Actions -->
          <div class="message-actions mt-2 pt-2 border-top border-light d-flex gap-2 justify-content-end" v-if="!msg.deleted">
             <div class="d-flex gap-2">
               <span v-for="emoji in commonEmojis" :key="emoji" @click="doComment(msg.id, emoji)" class="emoji-opt" style="font-size: 0.9rem;">{{ emoji }}</span>
             </div>
             <div class="ms-2 border-start ps-2 d-flex gap-2">
               <button @click="openForwardModal(msg.id)" class="btn btn-sm p-0 text-inherit opacity-75" title="Forward">
                 <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" fill="currentColor" viewBox="0 0 16 16">
                   <path d="M13.5 1a1.5 1.5 0 1 0 0 3 1.5 1.5 0 0 0 0-3zM11 2.5a2.5 2.5 0 1 1 .603 1.628l-6.718 3.12a2.499 2.499 0 0 1 0 1.504l6.718 3.12a2.5 2.5 0 1 1-.488.876l-6.718-3.12a2.5 2.5 0 1 1 0-3.256l6.718-3.12A2.5 2.5 0 0 1 11 2.5z"/>
                 </svg>
               </button>
               <button v-if="msg.senderId === myUsername" @click="doDelete(msg.id)" class="btn btn-sm p-0 text-danger opacity-75" title="Delete">
                 <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" fill="currentColor" viewBox="0 0 16 16">
                   <path d="M5.5 5.5A.5.5 0 0 1 6 6v6a.5.5 0 0 1-1 0V6a.5.5 0 0 1 .5-.5zm2.5 0a.5.5 0 0 1 .5.5v6a.5.5 0 0 1-1 0V6a.5.5 0 0 1 .5-.5zm3 .5a.5.5 0 0 0-1 0v6a.5.5 0 0 0 1 0V6z"/>
                   <path fill-rule="evenodd" d="M14.5 3a1 1 0 0 1-1 1H13v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V4h-.5a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1H6a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1h3.5a1 1 0 0 1 1 1v1zM4.118 4 4 4.059V13a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1V4.059L11.882 4H4.118zM2.5 3V2h11v1h-11z"/>
                 </svg>
               </button>
             </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Input Area -->
    <div class="input-area mt-auto border-top p-3 bg-white rounded-bottom">
      <div v-if="replyingTo" class="reply-preview alert alert-info py-2 px-3 mb-2 d-flex justify-content-between align-items-center shadow-sm">
        <div class="text-truncate">
          <strong>Replying to:</strong> {{ replyingTo.text }}
        </div>
        <button type="button" class="btn-close" @click="replyingTo = null"></button>
      </div>
      
      <div class="d-flex align-items-center gap-2">
        <textarea 
          class="form-control" 
          rows="1" 
          v-model="inputText" 
          @keydown="handleKeydown"
          placeholder="Type a message..."
          style="resize: none; border-radius: 20px;"
        ></textarea>
        <button 
          class="btn btn-primary d-flex align-items-center justify-content-center" 
          style="width: 40px; height: 40px; border-radius: 50%;"
          @click="doSend" 
          :disabled="!inputText.trim() || isLoading"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" fill="currentColor" viewBox="0 0 16 16">
            <path d="M15.854.146a.5.5 0 0 1 .11.54l-5.819 14.547a.75.75 0 0 1-1.329.124l-3.178-4.995L.643 7.184a.75.75 0 0 1 .124-1.33L15.314.037a.5.5 0 0 1 .54.11ZM6.636 10.07l2.761 4.338L14.13 2.576 6.636 10.07Zm6.787-8.201L1.591 6.602l4.339 2.76 7.494-7.493Z"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- Forward Dialog Overlay -->
    <div v-if="showForwardModal" class="forward-modal-overlay position-absolute top-0 start-0 w-100 h-100 d-flex align-items-center justify-content-center p-3" style="z-index: 1050; background: rgba(0,0,0,0.4);">
      <div class="card shadow w-100" style="max-width: 400px; max-height: 80%;">
        <div class="card-header bg-white d-flex justify-content-between align-items-center">
          <h5 class="mb-0">Forward message to...</h5>
          <button type="button" class="btn-close" @click="showForwardModal = false"></button>
        </div>
        <div class="card-body overflow-auto p-0">
          <div class="list-group list-group-flush">
            <button 
              v-for="conv in availableConversations" 
              :key="conv.id"
              class="list-group-item list-group-item-action py-3"
              @click="confirmForward(conv.id)"
            >
              <div class="fw-bold">{{ conv.name || conv.id }}</div>
              <small class="text-muted">{{ conv.isGroup ? 'Group' : 'Direct Chat' }}</small>
            </button>
            <div v-if="availableConversations.length === 0" class="p-4 text-center text-muted">
              No other conversations found.
            </div>
          </div>
        </div>
        <div class="card-footer bg-light text-end">
          <button class="btn btn-secondary btn-sm" @click="showForwardModal = false">Cancel</button>
        </div>
      </div>
    </div>
    
  </div>
</template>

<style scoped>
.conversation-container {
  background-color: #e5ddd5;
}

.messages-list {
  scrollbar-width: thin;
}

.bubble-right {
  border-bottom-right-radius: 2px !important;
}

.bubble-left {
  border-bottom-left-radius: 2px !important;
}

.emoji-opt {
  cursor: pointer;
  transition: transform 0.1s;
}

.emoji-opt:hover {
  transform: scale(1.4);
}

.message-bubble .message-actions {
  display: none;
}

.message-bubble:hover .message-actions {
  display: flex;
}
</style>
