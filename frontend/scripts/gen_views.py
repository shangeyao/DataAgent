# -*- coding: utf-8 -*-
import os

pages = {
    'Ingest': ('数据接入', '/catalog/tables', 'tables', '文件到表、离线/实时同步与数据源接入'),
    'Sources': ('数据源管理', '/sources', 'sources', '连接注册、连通性测试与关联任务'),
    'Studio': ('开发工作台', '/notebooks', 'notebooks', 'Notebook、SQL IDE 与 AI 代码助手'),
    'Flows': ('工作流编排', '/flows', 'flows', '可视化 DAG、依赖与调度配置'),
    'Cicd': ('发布管理', '/bundles', 'bundles', '配置即代码，Bundle 校验与多环境发布'),
    'Ops': ('运维监控', '/ops/overview', '', '四层运行视图、日志与 AI 诊断'),
    'Experiments': ('模型实验', '/models', 'experiments', '实验跟踪、指标对比与超参记录'),
    'Features': ('特征管理', '/features', 'features', '特征定义、物化计算与训练集构建'),
    'Registry': ('模型管理', '/registry/models', 'models', '模型注册、版本与血缘'),
    'Serving': ('模型服务', '/serving/services', 'services', '在线服务、副本与启停'),
    'Semantic': ('统一语义层', '/metrics', 'metrics', '本体建模、指标口径与维度定义'),
    'Ontology': ('本体建模', '/ontology/entities', 'entities', '实体关系图、业务术语表与指标口径'),
    'Quality': ('数据质量', '/quality/rules', 'rules', '规则引擎、执行结果与异常检测'),
    'Security': ('Agent 安全防护', '/security/policy', '', '注入检测、三要素拦截与权限绑定'),
    'Agents': ('智能体', '/agents', 'agents', '内置与自定义 Agent'),
    'Mcp': ('MCP 服务', '/settings/tools', 'tools', 'Model Context Protocol 工具暴露'),
    'ModelList': ('模型列表', '/settings/integrations', 'providers', '多模型接入与 Endpoint 管理'),
    'Compute': ('计算资源', '/compute-resources', 'types', '四类集群、CU 配额与网络设置'),
    'Members': ('成员与权限', '/users', 'users', '工作空间成员与角色分配'),
    'Settings': ('工作空间设置', '/settings', '', '通知渠道、Git 连接与偏好'),
    'Workspaces': ('工作空间', '/console/workspaces', 'workspaces', '管理工作空间及其资源状态'),
    'ConsoleUsers': ('用户管理', '/users', 'users', '控制台用户与全局角色'),
    'Platform': ('平台设置', '/settings', '', '模型接入、平台信息与审计'),
    'Dashboards': ('仪表盘', '/dashboards', 'dashboards', '看板搭建、卡片渲染与分享'),
    'Analysis': ('分析空间', '/analysis-spaces', 'spaces', '面向业务方的分析空间与资产绑定'),
}

out_dir = os.path.join(os.path.dirname(__file__), 'src', 'views')
for name, (title, endpoint, listkey, desc) in pages.items():
    lk = (' list-key="' + listkey + '"') if listkey else ''
    content = (
        '<script setup lang="ts">\n'
        "import GenericView from '../components/GenericView.vue'\n"
        '</script>\n\n'
        '<template>\n'
        f'  <GenericView title="{title}" endpoint="{endpoint}"{lk} description="{desc}" />\n'
        '</template>\n'
    )
    with open(os.path.join(out_dir, name + '.vue'), 'w', encoding='utf-8') as f:
        f.write(content)
print('regenerated', len(pages), 'views')
