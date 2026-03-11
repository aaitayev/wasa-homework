<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import api from '../services/axios.js';

const conversations = ref([]);
const isLoading = ref(false);
const errorMsg = ref('');
const router = useRouter();

async function fetchConversations() {
  errorMsg.value = '';
  isLoading.value = true;
  conversations.value = [];
  try {
    const res = await api.get('/conversations');
    const validDetails = res.data || [];
      
    // process into ui model
    conversations.value = validDetails.map(conv => {
      let title = conv.name || conv.id;
      if (!conv.isGroup && conv.participants) {
        const myUsername = localStorage.getItem('username');
        const others = conv.participants.filter(p => p !== myUsername);
        if (others.length > 0) {
          title = others.join(', ');
        }
      }
      
      const dt = conv.lastMessageAt ? new Date(conv.lastMessageAt) : new Date(0);
      
      return {
        id: conv.id,
        title: title,
        lastActivity: dt,
        snippet: conv.lastMessageText || 'No messages',
      };
    }); // already sorted backwards chronologically by backend
    
  } catch (error) {
    if (error.response && error.response.status === 401) {
      router.push('/login');
    } else {
      errorMsg.value = 'Failed to load conversations.';
      console.error(error);
    }
  } finally {
    isLoading.value = false;
  }
}

async function doNewChat() {
  const username = prompt("Enter username to chat with:");
  if (!username) return;
  
  try {
    isLoading.value = true;
    errorMsg.value = '';
    // Create new conversation
    await api.post('/messages', {
      text: "Started chat with " + username,
      isGroup: false
    });
    // Refresh to show newly joined conversation
    await fetchConversations();
  } catch (error) {
    errorMsg.value = 'Failed to create new chat.';
    console.error(error);
  } finally {
    isLoading.value = false;
  }
}

async function doNewGroup() {
  const groupName = prompt("Enter group name:");
  if (!groupName) return;
  
  const usernamesStr = prompt("Enter comma-separated usernames:");
  if (!usernamesStr) return;
  
  const usernames = usernamesStr.split(',').map(s => s.trim()).filter(s => s);
  
  try {
    isLoading.value = true;
    errorMsg.value = '';
    
    // 1. Create group via POST /messages
    const msgRes = await api.post('/messages', {
      text: "Group " + groupName + " created",
      isGroup: true
    });
    
    const groupId = msgRes.data.conversationId;
    if (!groupId) throw new Error("No groupId returned");
    
    // 2. setGroupName
    await api.put(`/groups/${groupId}/name`, { name: groupName });
    
    // 3. addToGroup for each member
    for (const member of usernames) {
      try {
        await api.post(`/groups/${groupId}/members`, { memberId: member });
      } catch (err) {
        console.warn("Failed to add member", member, err);
      }
    }
    
    await fetchConversations();
  } catch (error) {
    errorMsg.value = 'Failed to create new group.';
    console.error(error);
  } finally {
    isLoading.value = false;
  }
}

function openConversation(id) {
  router.push('/conversations/' + id);
}

function formatTime(dateObj) {
  if (!dateObj || dateObj.getTime() === 0) return '';
  return dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

onMounted(() => {
  fetchConversations();
});
</script>

<template>
  <div>
    <div class="d-flex justify-content-between align-items-center mb-3 border-bottom pb-2 mt-3">
      <h2>Conversations</h2>
      <div>
        <button class="btn btn-outline-secondary btn-sm me-2" @click="fetchConversations">Refresh</button>
        <button class="btn btn-outline-primary btn-sm me-2" @click="doNewChat">New chat</button>
        <button class="btn btn-outline-success btn-sm" @click="doNewGroup">New group</button>
      </div>
    </div>
    
    <ErrorMsg v-if="errorMsg" :msg="errorMsg" />
    <LoadingSpinner v-if="isLoading" />
    
    <div class="list-group" v-if="!isLoading && conversations.length > 0">
      <button 
        v-for="conv in conversations" 
        :key="conv.id" 
        class="list-group-item list-group-item-action py-3 lh-sm"
        @click="openConversation(conv.id)"
      >
        <div class="d-flex w-100 align-items-center justify-content-between mb-1">
          <strong class="mb-1 text-truncate">{{ conv.title }}</strong>
          <small class="text-nowrap text-muted ms-2">{{ formatTime(conv.lastActivity) }}</small>
        </div>
        <div class="col-10 mb-1 small text-muted text-truncate">{{ conv.snippet }}</div>
      </button>
    </div>
    
    <div v-else-if="!isLoading && conversations.length === 0" class="text-center text-muted my-5">
      <p>No conversations found.</p>
      <p>Click "New chat" or "New group" to start messaging.</p>
    </div>
  </div>
</template>

<style scoped>
</style>
