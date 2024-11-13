<template>
  <div class="app" id="app">
    <header class="header">
      <input v-model.number="inputId" placeholder="Enter LiveId..." class="input"/>
      <button @click="connect" class="button">连接</button>
      <button @click="disconnect" class="button">断开</button>
    </header>
    <main class="main">
      <pre ref="output" class="output">{{ logs }}</pre>
    </main>
    <div v-if="message" class="message">{{ message }}</div> <!-- 显示提示信息 -->
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue';
import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime.js";
import { Start, Shutdown } from "../../wailsjs/go/app/App.js";

const inputId = ref(null); // 输入框内容
const logs = ref("");
const message = ref(""); // 输出框内容
const maxLines = 500; // 最多保存的行数

const connect = async () => {
  try {
    const id = parseInt(inputId.value);
    if (!isNaN(id)) {
      // 调用 Go 的 NewView 函数，接收返回信息
      message.value = await Start(id);
      updateLog(); // 显示返回的提示信息
    }
  } catch (error) {
    message.value += `Error connecting: ${error}\n`;
    updateLog();
  }
};

const disconnect = async () => {
  Shutdown();
  message.value = "连接已断开";
};

const updateLog = () => {
  nextTick(() => {
    const logOutput = document.querySelector('.output');
    logOutput.scrollTop = logOutput.scrollHeight;
    const lines = logs.value.split("\n");
    if (lines.length > maxLines) {
      logs.value = lines.slice(-maxLines).join("\n");
    }
  });
};

const appendOutput = (output) => {
  logs.value += output + "\n";
  updateLog();
};

onMounted(() => {
  // 监听 Go 的输出事件
  EventsOn("new-output", (output) => {
    // 按行追加新数据
    appendOutput(output);
  });
});

onBeforeUnmount(() => {
  // 移除事件监听器
  EventsOff("new-output");
});
</script>

<style scoped>
#app {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-image: linear-gradient(-20deg, #e9defa 0%, #fbfcdb 100%);
  backdrop-filter: blur(10px);
}
.app {
  box-shadow: rgba(0, 0, 0, 0.16) 0px 3px 6px, rgba(0, 0, 0, 0.23) 0px 3px 6px;
  border-radius: 8px;
}

.header {
  display: flex;
  align-items: center;
  padding: 10px;
}
.header .input {
  margin-left: 10px;
}
.header .button {
  margin-left: 10px;
}
.main {
  flex-grow: 1;
  overflow: auto;
  border-top: 1px solid #CCC;
}
.main .output {
  white-space: pre-wrap;
  word-wrap: break-word;
  height: 100%;
  overflow-y: scroll;
  text-align: left;
  padding-left: 20px;
  color: black;
}
.main .output::-webkit-scrollbar {
  width: 0 !important;
}
.message {
  color: black;
}
</style>
