<template>
  <el-dialog :visible.sync="dialogVisible" :title="nameStatus">
    <el-form
      ref="jobEditForm"
      :model="jobConfig"
      :rules="addJobRules"
      label-width="80px "
    >
      <el-form-item label="系统" prop="system">
        <el-select v-model="jobConfig.system" placeholder="任务所属系统">
          <el-option
            v-for="(desc, name) in systems"
            :key="name"
            :label="desc + ' - ' + name"
            :value="name"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="任务名称" prop="name">
        <el-input
          v-model="jobConfig.name"
          placeholder="任务名称， 限制只允许： a-z A-Z 0-9 - _，例如：testJob"
        />
      </el-form-item>
      <el-form-item label="任务说明" prop="desc">
        <el-input
          v-model="jobConfig.desc"
          placeholder="topic 中文说明，尽量简洁"
        />
      </el-form-item>
      <el-form-item label="任务并发" prop="concurrency">
        <el-input-number
          v-model="jobConfig.concurrency"
          :min="1"
          controls-position="right"
        />
      </el-form-item>
      <el-form-item label="超时时间">
        <el-input
          v-model="jobConfig.timeout"
          placeholder="任务执行超时时间，单位秒，默认 300 秒"
        />
      </el-form-item>
      <el-form-item label="环境变量">
        <template v-for="(env, index) in envs">
          <el-row :gutter="5" :key="index" type="flex">
            <el-col :span="7">
              <el-input v-model="env.key" placeholder="Key" clearable />
            </el-col>
            <el-col :span="1">=</el-col>
            <el-col :span="12">
              <el-input v-model="env.value" placeholder="Value" clearable />
            </el-col>
            <el-col :span="4">
              <el-row :gutter="5">
                <el-col :span="10">
                  <el-button
                    :disabled="envs.length == 1"
                    type="danger"
                    icon="el-icon-delete"
                    circle
                    @click="delEnv(index)"
                  />
                </el-col>
                <el-col v-if="index == envs.length - 1" :span="10">
                  <el-button
                    type="primary"
                    icon="el-icon-plus"
                    circle
                    @click="addEnv"
                  />
                </el-col>
              </el-row>
            </el-col>
          </el-row>
        </template>
      </el-form-item>
      <el-form-item
        :label-width="formLabelWidth"
        label="执行器"
        prop="executor"
      >
        <el-select v-model="jobConfig.executor" placeholder="请选择执行器">
          <el-option key="0" :value="0" label="Shell 执行器" />
          <el-option key="1" :value="1" label="GRPC 执行器" />
        </el-select>
      </el-form-item>
      <el-form-item
        v-if="jobConfig.executor === 0"
        label="执行命令"
        prop="command"
      >
        <el-input
          v-model="jobConfig.command"
          type="textarea"
          rows="2"
          placeholder="shell 任务执行命令"
        />
      </el-form-item>
      <el-form-item
        v-if="jobConfig.executor === 1"
        label="Host"
        prop="grpc_host"
      >
        <el-input
          v-model="jobConfig.grpc_host"
          type="textarea"
          rows="1"
          placeholder="GRPC 服务器地址 Host:Port"
        />
      </el-form-item>
      <el-form-item label="依赖任务">
        <el-tag
          v-for="(dep, index) in dependents"
          :key="dep.value"
          :class="index == 0 ? '' : 'tag-dep'"
          closable
          @close="delDependentJob(index, false)"
        >
          {{ dep.label }}
        </el-tag>
        <el-cascader
          v-model="addDependentInput"
          :options="allJobs"
          :show-all-levels="false"
          size="small"
          class="button-add-dep"
          placeholder="+ 添加依赖"
          @change="dependentAdded(false)"
        />
      </el-form-item>
      <el-form-item label="子任务">
        <el-tag
          v-for="(child, index) in children"
          :key="child.value"
          :class="index == 0 ? '' : 'tag-dep'"
          closable
          @close="delDependentJob(index, true)"
        >
          {{ child.label }}
        </el-tag>
        <el-cascader
          v-model="addDependentInput"
          :options="allJobs"
          :show-all-levels="false"
          size="small"
          class="button-add-dep"
          placeholder="+ 添加子任务"
          @change="dependentAdded(true)"
        />
      </el-form-item>
      <el-form-item
        v-if="dependents.length == 0"
        label="执行方式"
        prop="scheduler_val"
      >
        <el-tooltip placement="top">
          <div slot="content">{{ jobDesc }}</div>
          <el-input
            v-model="scheduler.value"
            :placeholder="
              scheduler.type == 0 ? '* * * * * (分 时 天 月 周)' : '执行间隔'
            "
          >
            <el-select
              slot="prepend"
              v-model="scheduler.type"
              style="width: 100px"
              @change="schedulerTypeChanged"
            >
              <el-option label="@every" value="1" />
              <el-option label="crontab" value="0" />
              <el-option label="infinite" value="2" />
            </el-select>
            <el-select
              v-if="scheduler.type == 1 || scheduler.type == '@every'"
              slot="append"
              v-model="scheduler.unit"
              style="width: 80px"
            >
              <el-option label="分钟" value="m" />
              <el-option label="小时" value="h" />
              <el-option label="天" value="d" />
            </el-select>
            <el-select
              v-if="scheduler.type == 2 || scheduler.type == 'infinite'"
              slot="append"
              v-model="scheduler.unit"
              style="width: 80px"
            >
              <el-option label="毫秒" value="ms" />
              <el-option label="秒" value="s" />
              <el-option label="分钟" value="m" />
            </el-select>
          </el-input>
        </el-tooltip>
      </el-form-item>
    </el-form>
    <div slot="footer" class="dialog-footer">
      <el-button @click="dialogVisible = false">取 消</el-button>
      <el-button type="primary" @click="saveJob">确 定</el-button>
    </div>
  </el-dialog>
</template>

<script>
import { addJob, updateJob } from "@/api/job";

export default {
  props: {
    systems: {
      type: Object,
      default: () => {
        return {};
      }
    },
    jobs: {
      type: Array,
      default: () => {
        return [];
      }
    },
    toModifyItem: {
      type: Object,
      default: () => {
        return {
          system: ""
        };
      }
    },
    job: {
      type: Object,
      default: () => {}
    }
  },

  data() {
    return {
      nameStatus: "添加任务",
      dialogVisible: false,
      jobConfig: {
        system: "",
        name: "",
        desc: "",
        timeout: "",
        executor: 0,
        grpc_host: "",
        command: "",
        concurrency: 1
      },
      envs: [{ key: "", value: "" }],
      dependents: [],
      children: [],
      addDependentInput: [],
      scheduler: {
        value: "",
        type: "1",
        unit: "m"
      },
      addJobRules: {
        system: [{ required: true, message: "请选择系统", trigger: "blur" }],
        name: [
          { required: true, message: "请填写任务名称", trigger: "blur" },
          { validator: this.validateName, trigger: "blur" }
        ],
        desc: [{ required: true, message: "请填写任务描述", trigger: "blur" }],
        concurrency: [
          { required: true, message: "请选择任务并发数", trigger: "blur" }
        ]
        // command: [
        // { required: true, message: "请输入 shell 执行命令", trigger: "blur" }
        // ]
        // scheduler_val: [{ required: true, message: '请输入执行方式', trigger: 'blur' }]
      }
    };
  },

  computed: {
    allJobs: function() {
      var ret = [];
      var system = {};
      this.jobs.forEach(job => {
        if (system[job.system] === undefined) {
          system[job.system] = {
            value: job.system,
            label: job.system,
            children: []
          };
        }
        system[job.system].children.push({
          value: job.id,
          label: job.name
        });
      });
      for (var sys in system) {
        ret.push(system[sys]);
      }
      return ret;
    },
    jobDesc: function() {
      var type = parseInt(this.scheduler.type);
      if (type === 0) {
        return "crontab: 与 Linux 用法一致";
      } else if (type === 1) {
        return "@every: 以指定间隔执行任务";
      } else if (type === 2) {
        return "infinite: 持续执行，每两次执行 Sleep 间隔";
      }
    }
  },

  methods: {
    validateName(rule, value, callback) {
      var str = String(value);
      var reg = new RegExp(/([a-z,A-Z,0-9,_]+)/);
      var rstr = str.replace(reg, "");
      if (rstr === null || rstr === "") {
        return callback();
      } else {
        return callback(new Error("包含不合法字符 " + rstr));
      }
    },
    addEnv() {
      this.envs.push({ key: "", value: "" });
    },
    delEnv(index) {
      this.envs.splice(index, 1);
    },
    schedulerTypeChanged(val) {
      val = parseInt(val);
      if (val === 1) {
        this.scheduler.unit = "m";
      } else if (val === 2) {
        this.scheduler.unit = "s";
      }
    },
    dependentAdded(child) {
      if (this.addDependentInput[1]) {
        this.allJobs.forEach(system => {
          if (system.value === this.addDependentInput[0]) {
            system.children.forEach(job => {
              if (job.value === this.addDependentInput[1]) {
                job.disabled = true;
                job.system = this.addDependentInput[0];
                if (child) {
                  this.children.push(job);
                } else {
                  this.dependents.push(job);
                }
              }
            });
          }
        });
      }
      this.addDependentInput = [];
    },
    delDependentJob(index, child) {
      var jobId = "";
      if (child) {
        jobId = this.children[index].value;
        this.children.splice(index, 1);
      } else {
        jobId = this.dependents[index].value;
        this.dependents.splice(index, 1);
      }
      this.allJobs.forEach(system => {
        system.children.forEach(job => {
          if (job.value === jobId) {
            job.disabled = false;
          }
        });
      });
    },
    resetJobForm() {
      this.nameStatus = "添加任务";
      this.jobConfig = {
        system: "",
        name: "",
        desc: "",
        executor: 0,
        grpc_host: "",
        timeout: "",
        command: "",
        concurrency: 1
      };
      this.envs = [{ key: "", value: "" }];
      this.dependents = [];
      this.children = [];
      this.addDependentInput = [];
      this.scheduler = {
        value: "",
        type: "1",
        unit: "m"
      };
      this.$nextTick(() => {
        this.$refs["jobEditForm"].clearValidate();
      });
    },

    setjobForm(job) {
      this.nameStatus = "修改任务";
      // this.jobConfig.id = job.id
      this.jobConfig = {
        id: job.id,
        system: job.system,
        create_at: job.create_at,
        name: job.name,
        desc: job.desc,
        executor: job.executor,
        grpc_host: job.grpc_host,
        timeout: job.timeout,
        command: job.command,
        concurrency: job.concurrency
      };
      this.envs = JSON.parse(job.envs);
      if (this.envs.length === 0) {
        this.envs = [{ key: "", value: "" }];
      }
      this.dependents = JSON.parse(job.dependents);

      this.children = JSON.parse(job.children);
      this.addDependentInput = [];
      this.scheduler = JSON.parse(job.scheduler);
      this.scheduler.type = this.scheduler.type.toString();

      this.$nextTick(() => {
        this.$refs["jobEditForm"].clearValidate();
      });
    },
    saveJob() {
      this.$refs["jobEditForm"].validate(valid => {
        if (!valid) {
          return;
        }
        this.jobConfig.timeout = parseInt(this.jobConfig.timeout);

        if (this.scheduler.type === "@every") {
          this.scheduler.type = 1;
        }
        if (this.scheduler.type === "crontab") {
          this.scheduler.type = 0;
        }
        if (this.scheduler.type === "infinite") {
          this.scheduler.type = 2;
        }
        this.scheduler.type = parseInt(this.scheduler.type);
        if (this.scheduler.type === 0) {
          var crontab = this.scheduler.value.trim().split(" ");
          if (crontab.length === 5) {
            this.scheduler.value = 0 + " " + this.scheduler.value;
          }
        }
        this.jobConfig.scheduler = JSON.stringify(this.scheduler);
        this.jobConfig.dependents = JSON.stringify(this.dependents);
        this.jobConfig.children = JSON.stringify(this.children);
        var envs = [];
        this.envs.forEach(env => {
          if (env.key !== "") {
            envs.push(env);
          }
        });
        this.jobConfig.envs = JSON.stringify(envs);
        if (this.jobConfig.id === undefined || this.jobConfig.id === null) {
          addJob(this.jobConfig).then(res => {
            if (res.code === 0) {
              this.$notify({
                title: "成功",
                message: "保存成功",
                type: "success",
                duration: 2000
              });
              this.dialogVisible = false;
              this.$emit("jobDialogAsk", "saved");
            }
          });
        } else {
          updateJob(this.jobConfig).then(res => {
            if (res.code === 0) {
              this.$notify({
                title: "成功",
                message: "保存成功",
                type: "success",
                duration: 2000
              });
              this.dialogVisible = false;
              this.$emit("jobDialogAsk", "saved");
            }
          });
        }
      });
    }
  }
};
</script>

<style lang="scss" scoped>
.el-row {
  margin-bottom: 5px;

  &:last-child {
    margin-bottom: 0;
  }
}
.tag-dep {
  margin-left: 5px;
}
.button-add-dep {
  width: 130px;
}
</style>
