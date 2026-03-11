<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import api, { setToken } from '../services/axios.js';

const username = ref('');
const errorMsg = ref('');
const isLoading = ref(false);
const router = useRouter();

async function doLogin() {
  if (username.value.length < 3) {
    errorMsg.value = "Username must be at least 3 characters long.";
    return;
  }
  
  errorMsg.value = '';
  isLoading.value = true;
  
  try {
    const response = await api.post('/session', {
      name: username.value
    });
    
    // Server returns 201 with body { "identifier": "some_token" }
    const token = response.data.identifier;
    if (token) {
      setToken(token);
      localStorage.setItem('username', username.value);
      router.push('/conversations');
    } else {
      errorMsg.value = "Login failed: No identifier returned.";
    }
  } catch (error) {
    if (error.response && error.response.data && error.response.data.message) {
      errorMsg.value = `Login failed: ${error.response.data.message}`;
    } else {
      errorMsg.value = "An error occurred during login. Please try again.";
    }
  } finally {
    isLoading.value = false;
  }
}
</script>

<template>
  <div class="row justify-content-center mt-5">
    <div class="col-md-6 col-lg-4">
      <div class="card shadow-sm">
        <div class="card-body p-4">
          <h2 class="card-title text-center mb-4">Login</h2>
          
          <ErrorMsg v-if="errorMsg" :msg="errorMsg" class="mb-3"></ErrorMsg>
          
          <form @submit.prevent="doLogin">
            <div class="mb-3">
              <label for="usernameInput" class="form-label">Username</label>
              <input 
                type="text" 
                class="form-control" 
                id="usernameInput" 
                v-model.trim="username" 
                placeholder="Enter username" 
                required 
                minlength="3" 
                autocomplete="username"
              >
              <div class="form-text">Username must be at least 3 characters.</div>
            </div>
            
            <div class="d-grid mt-4">
              <button 
                type="submit" 
                class="btn btn-primary" 
                :disabled="isLoading || username.length < 3"
              >
                <LoadingSpinner v-if="isLoading" class="me-2" />
                Login
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
