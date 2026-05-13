<template>
  <div class="ai-qa-view">
    <div class="mb-6">
      <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100">AI 问答</h2>
      <p class="text-gray-500 dark:text-gray-400 text-sm mt-1">
        选择一份或多份已导入文档后，基于文档内容进行问答。
      </p>
    </div>

    <AnalysisHistoryQaPanel />
  </div>
</template>

<script setup lang="ts">
import AnalysisHistoryQaPanel from '@/components/AnalysisHistoryQaPanel.vue'
</script>
