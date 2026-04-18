import { Button, Card, Chip, Form, Input, Label, Modal, Switch, TextField, Tooltip, toast } from "@heroui/react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Copyable, CopyButton } from "@/components/CopyButton";
import { EyeToggle, maskKey } from "@/components/MaskedKey";
import { TagInput } from "@/components/TagInput";
import type { Rule } from "@/utils/api";
import * as api from "@/utils/api";
import { useAuth } from "@/utils/auth";
import { generateKey } from "@/utils/key";
import { useConfirmAction } from "@/utils/useConfirmAction";

function RefreshKeyButton({ onRefresh }: { onRefresh: () => void }) {
  const { pending, trigger } = useConfirmAction(onRefresh, 3_000);
  return (
    <Tooltip delay={0} closeDelay={0} shouldCloseOnPress={false}>
      <Tooltip.Trigger className="w-4 h-4">
        <button
          type="button"
          className="inline-flex items-center justify-center w-4 h-4 text-muted hover:text-foreground cursor-pointer transition-colors"
          onClick={trigger}
          aria-label={pending ? "确认刷新密钥" : "刷新密钥"}
        >
          <span className={`icon-[mdi--refresh] w-3.5 h-3.5 ${pending ? "text-accent" : "text-muted"}`} />
        </button>
      </Tooltip.Trigger>
      <Tooltip.Content>{pending ? "确认刷新密钥" : "刷新密钥"}</Tooltip.Content>
    </Tooltip>
  );
}

function ConfirmDeleteButton({ onDelete, isDisabled }: { onDelete: () => void; isDisabled?: boolean }) {
  const { pending, trigger } = useConfirmAction(onDelete, 3_000);
  return (
    <Button variant="danger" size="sm" className="h-7 text-xs px-2" onPress={trigger} isDisabled={isDisabled}>
      <span className="icon-[mdi--delete-outline] w-3 h-3" />
      {pending ? "确认删除" : "删除"}
    </Button>
  );
}

type RuleDraft = Omit<Rule, "id">;

const emptyDraft: RuleDraft = {
  name: "",
  src: "",
  dest: "",
  dest_api_key: "",
  skip_cert_verify: true,
  api_key: "",
  force_api_key: true,
  comment: "",
  tags: [],
};

function KeyMapping({
  src = "",
  dest = "",
  one,
  onRefresh,
}: {
  src?: string;
  dest?: string;
  one?: string;
  onRefresh?: () => void;
}) {
  const [visible, setVisible] = useState(false);
  const hasSrc = !!src;
  const hasDest = !!dest;
  const isOne = typeof one === "string";
  const hasOne = !!one;
  const hasAny = isOne ? hasOne : hasSrc || hasDest;

  if (!hasAny) {
    return <span className="text-muted text-xs">无</span>;
  }

  return (
    <div className="flex items-center gap-1.5">
      {isOne ? (
        <Copyable text={src}>
          <span className="px-1.5 py-0.5 bg-surface-secondary rounded text-foreground text-xs break-all cursor-pointer transition-colors hover:bg-border">
            {visible ? one : maskKey(one)}
          </span>
        </Copyable>
      ) : (
        <>
          <Copyable text={src}>
            <span className="px-1.5 py-0.5 bg-surface-secondary rounded text-foreground text-xs break-all cursor-pointer transition-colors hover:bg-border">
              {hasSrc ? (visible ? src : maskKey(src)) : "(空)"}
            </span>
          </Copyable>
          <span className="text-border">→</span>
          <Copyable text={dest}>
            <span className="px-1.5 py-0.5 bg-surface-secondary rounded text-foreground text-xs break-all cursor-pointer transition-colors hover:bg-border">
              {hasDest ? (visible ? dest : maskKey(dest)) : "(空)"}
            </span>
          </Copyable>
        </>
      )}
      <span className="flex-1" />
      <span className="flex items-center justify-center gap-2 shrink-0">
        <EyeToggle visible={visible} onToggle={() => setVisible((v) => !v)} />
        {onRefresh ? <RefreshKeyButton onRefresh={onRefresh} /> : null}
      </span>
    </div>
  );
}

function FormFields({
  id,
  data,
  onChange,
  onSubmit,
  showRefreshKey,
  forbidEnterSubmit = false,
}: {
  id: string;
  data: RuleDraft;
  onChange: (patch: Partial<RuleDraft>) => void;
  onSubmit: () => void;
  showRefreshKey?: boolean;
  forbidEnterSubmit?: boolean;
}) {
  return (
    // biome-ignore lint: wrapper div for enter key prevention
    <div
      onKeyDown={
        forbidEnterSubmit
          ? (e) => {
              if (e.key === "Enter" && (e.target as HTMLElement).tagName !== "BUTTON") e.preventDefault();
            }
          : undefined
      }
    >
      <Form
        id={id}
        className="flex flex-col gap-4"
        validationBehavior="native"
        onSubmit={(e) => {
          e.preventDefault();
          onSubmit();
        }}
      >
        <TextField isRequired value={data.name} onChange={(v) => onChange({ name: v })}>
          <Label>名称</Label>
          <Input placeholder="规则名称" />
        </TextField>
        <TextField isRequired value={data.src} onChange={(v) => onChange({ src: v })}>
          <Label>源路径</Label>
          <Input placeholder="/openai/v1" />
        </TextField>
        <TextField isRequired value={data.dest} onChange={(v) => onChange({ dest: v })}>
          <Label>目标地址</Label>
          <Input placeholder="https://api.openai.com/v1" />
        </TextField>
        <TextField value={data.dest_api_key} onChange={(v) => onChange({ dest_api_key: v })}>
          <Label>目标 API Key</Label>
          <Input placeholder="目标服务密钥" />
        </TextField>
        <TextField value={data.api_key} onChange={(v) => onChange({ api_key: v })}>
          <div className="flex items-center justify-between">
            <Label>API Key</Label>
            {showRefreshKey ? (
              <button
                type="button"
                className="inline-flex items-center gap-1 text-xs text-accent hover:text-accent-hover cursor-pointer"
                onClick={() => onChange({ api_key: generateKey() })}
              >
                <span className="icon-[mdi--refresh] w-3.5 h-3.5" />
                生成新 Key
              </button>
            ) : null}
          </div>
          <Input placeholder="来源请求密钥" />
        </TextField>
        <div className="flex flex-col gap-1">
          <Label>标签</Label>
          <TagInput value={data.tags ?? []} onChange={(tags) => onChange({ tags })} placeholder="输入标签后回车" />
        </div>
        <TextField value={data.comment} onChange={(v) => onChange({ comment: v })}>
          <Label>备注</Label>
          <Input placeholder="业务说明" />
        </TextField>
        <Switch isSelected={data.force_api_key} onChange={() => onChange({ force_api_key: !data.force_api_key })}>
          <Switch.Control>
            <Switch.Thumb />
          </Switch.Control>
          <Label className="cursor-pointer">强制匹配 API Key</Label>
        </Switch>
        <Switch
          isSelected={data.skip_cert_verify}
          onChange={() => onChange({ skip_cert_verify: !data.skip_cert_verify })}
        >
          <Switch.Control>
            <Switch.Thumb />
          </Switch.Control>
          <Label className="cursor-pointer">忽略证书检查</Label>
        </Switch>
      </Form>
    </div>
  );
}

export default function RulesPage() {
  const { user } = useAuth();
  const isAdmin = user?.role === "admin";

  const [rows, setRows] = useState<Rule[]>([]);
  const [draft, setDraft] = useState(emptyDraft);
  const [createOpen, setCreateOpen] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editDraft, setEditDraft] = useState<RuleDraft | null>(null);
  const [loading, setLoading] = useState(false);
  const [proxyEndpoint, setProxyEndpoint] = useState("");

  const isEditing = useMemo(() => editingId !== null && editDraft !== null, [editingId, editDraft]);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      setRows(await api.listRules());
    } catch (err) {
      toast.danger(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  useEffect(() => {
    let cancelled = false;
    api
      .getInfo()
      .then(({ endpoint }) => {
        if (!cancelled) setProxyEndpoint(endpoint);
      })
      .catch((err) => {
        if (!cancelled) toast.danger(err instanceof Error ? err.message : String(err));
      });
    return () => {
      cancelled = true;
    };
  }, []);

  function beginEdit(row: Rule) {
    setEditingId(row.id);
    setEditDraft({
      name: row.name,
      src: row.src,
      dest: row.dest,
      dest_api_key: row.dest_api_key,
      skip_cert_verify: row.skip_cert_verify,
      api_key: row.api_key,
      force_api_key: row.force_api_key,
      comment: row.comment,
      tags: row.tags ?? [],
    });
  }

  function beginCreate() {
    setDraft(emptyDraft);
    setCreateOpen(true);
  }

  function cancelCreate() {
    setCreateOpen(false);
  }
  function cancelEdit() {
    setEditingId(null);
    setTimeout(() => setEditDraft(null), 350);
  }

  async function doCreate() {
    const { id } = await api.createRule(draft);
    setCreateOpen(false);
    const name = draft.name;
    setDraft(emptyDraft);
    await loadData();
    toast.success(`新增成功 #${id} ${name}`);
  }

  async function doSave() {
    if (editingId === null || editDraft === null) return;
    await api.updateRule({ id: editingId, ...editDraft });
    toast.success(`已保存 #${editingId} ${editDraft.name || ""}`);
    setEditingId(null);
    setEditDraft(null);
    await loadData();
  }

  async function doDelete(id: number) {
    await api.deleteRule(id);
    setRows((prev) => prev.filter((item) => item.id !== id));
    const deleted = rows.find((r) => r.id === id);
    toast.success(`已删除 #${id} ${deleted?.name || ""}`);
  }

  async function refreshApiKey(id: number) {
    const row = rows.find((r) => r.id === id);
    if (!row) return;
    const newKey = generateKey();
    await api.updateRule({ ...row, api_key: newKey });
    await loadData();
    toast.success(`已刷新 API Key`);
  }

  return (
    <>
      <main className="max-w-5xl mx-auto px-4 py-6">
        <div className="flex items-center justify-between mb-3">
          <p className="uppercase tracking-wide">
            <span className="text-foreground text-xl font-semibold">转发规则</span>
            <span className="text-muted text-xs">（共 {rows.length} 条）</span>
          </p>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onPress={() => void loadData()} isDisabled={loading}>
              <span className="icon-[mdi--refresh] w-4 h-4" />
              {loading ? "加载中..." : "刷新"}
            </Button>
            {isAdmin ? (
              <Button variant="primary" size="sm" onPress={beginCreate} isDisabled={isEditing || createOpen}>
                <span className="icon-[mdi--plus] w-4 h-4" />
                新增规则
              </Button>
            ) : null}
          </div>
        </div>

        <div className="flex items-center gap-2 text-sm text-muted mb-4">
          <p>
            <span className="font-semibold">转发端点：</span>
            <span>{proxyEndpoint || "加载中..."}</span>
          </p>
          {proxyEndpoint ? <CopyButton text={proxyEndpoint} /> : null}
        </div>

        {rows.length === 0 ? (
          <Card>
            <Card.Content className="p-6">
              <p className="text-center text-muted">暂无规则，请先新增一条。</p>
            </Card.Content>
          </Card>
        ) : (
          <div className="flex flex-col gap-2">
            {rows.map((row) => (
              <Card key={row.id} className="rounded-lg">
                <Card.Content className="px-1.5 py-0">
                  {/* Header: title + tags + actions */}
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex items-center gap-1.5 min-w-0 flex-wrap">
                      <span className="font-semibold text-sm">
                        #{row.id} {row.name || "-"}
                      </span>
                      {row.force_api_key ? (
                        <Chip size="sm" color="warning" variant="soft">
                          强制匹配
                        </Chip>
                      ) : null}
                      {row.skip_cert_verify ? (
                        <Chip size="sm" color="warning" variant="soft">
                          跳过证书
                        </Chip>
                      ) : null}
                      {(row.tags ?? []).map((tag) => (
                        <Chip key={tag} size="sm" color="accent" variant="secondary">
                          {tag}
                        </Chip>
                      ))}
                    </div>
                    {isAdmin ? (
                      <div className="flex items-center gap-1.5 shrink-0">
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-7 text-xs px-2"
                          onPress={() => beginEdit(row)}
                          isDisabled={isEditing || createOpen}
                        >
                          <span className="icon-[mdi--pencil-outline] w-3 h-3" />
                          编辑
                        </Button>
                        <ConfirmDeleteButton
                          onDelete={() =>
                            void doDelete(row.id).catch((err) =>
                              toast.danger(err instanceof Error ? err.message : String(err))
                            )
                          }
                          isDisabled={isEditing || createOpen}
                        />
                      </div>
                    ) : null}
                  </div>

                  {isAdmin ? (
                    <div className="grid grid-cols-[auto_1fr] gap-x-2 gap-y-1 text-sm mt-1">
                      <span className="text-muted text-xs self-center">路径映射</span>
                      <div className="flex items-center gap-1.5">
                        <Copyable text={row.src}>
                          <span className="px-1.5 py-0.5 bg-surface-secondary rounded text-foreground text-xs break-all cursor-pointer transition-colors hover:bg-border">
                            {row.src || "N/A"}
                          </span>
                        </Copyable>
                        <span className="text-border">→</span>
                        <Copyable text={row.dest}>
                          <span className="px-1.5 py-0.5 bg-surface-secondary rounded text-foreground text-xs break-all cursor-pointer transition-colors hover:bg-border">
                            {row.dest || "N/A"}
                          </span>
                        </Copyable>
                      </div>

                      <span className="text-muted text-xs self-center">密钥映射</span>
                      <KeyMapping
                        src={row.api_key}
                        dest={row.dest_api_key}
                        onRefresh={() =>
                          void refreshApiKey(row.id).catch((err) =>
                            toast.danger(err instanceof Error ? err.message : String(err))
                          )
                        }
                      />

                      {row.comment ? (
                        <>
                          <span className="text-muted text-xs self-center">备注</span>
                          <span className="text-muted text-xs">{row.comment}</span>
                        </>
                      ) : null}
                    </div>
                  ) : (
                    <div className="grid grid-cols-[auto_1fr] gap-x-2 gap-y-1 text-sm mt-1">
                      <span className="text-muted text-xs self-center">路径</span>
                      <Copyable text={row.src}>
                        <span className="px-1.5 py-0.5 bg-surface-secondary rounded text-foreground text-xs break-all cursor-pointer transition-colors hover:bg-border">
                          {row.src || "N/A"}
                        </span>
                      </Copyable>

                      <span className="text-muted text-xs self-center">密钥</span>
                      <KeyMapping one={row.api_key} />

                      {row.comment ? (
                        <>
                          <span className="text-muted text-xs self-center">备注</span>
                          <span className="text-muted text-xs">{row.comment}</span>
                        </>
                      ) : null}
                    </div>
                  )}
                </Card.Content>
              </Card>
            ))}
          </div>
        )}
      </main>

      {/* Create modal */}
      <Modal>
        <Modal.Backdrop
          isOpen={createOpen}
          onOpenChange={(open) => {
            if (!open) cancelCreate();
          }}
        >
          <Modal.Container>
            <Modal.Dialog>
              <Modal.CloseTrigger />
              <Modal.Header>
                <Modal.Heading>新增转发规则</Modal.Heading>
              </Modal.Header>
              <Modal.Body>
                <FormFields
                  id="form-create-rule"
                  data={draft}
                  onChange={(patch) => setDraft((prev) => ({ ...prev, ...patch }))}
                  onSubmit={() => {
                    void doCreate().catch((err) => toast.danger(err instanceof Error ? err.message : String(err)));
                  }}
                  showRefreshKey
                  forbidEnterSubmit
                />
              </Modal.Body>
              <Modal.Footer>
                <Button variant="secondary" slot="close">
                  取消
                </Button>
                <Button type="submit" form="form-create-rule" variant="primary">
                  提交新增
                </Button>
              </Modal.Footer>
            </Modal.Dialog>
          </Modal.Container>
        </Modal.Backdrop>
      </Modal>

      {/* Edit modal */}
      <Modal>
        <Modal.Backdrop
          isOpen={isEditing}
          onOpenChange={(open) => {
            if (!open) cancelEdit();
          }}
        >
          <Modal.Container>
            <Modal.Dialog>
              <Modal.CloseTrigger />
              <Modal.Header>
                <Modal.Heading>
                  编辑规则 #{editingId} {editDraft?.name || ""}
                </Modal.Heading>
              </Modal.Header>
              <Modal.Body>
                {editDraft !== null ? (
                  <FormFields
                    id="form-edit"
                    data={editDraft}
                    onChange={(patch) => setEditDraft((prev) => (prev ? { ...prev, ...patch } : prev))}
                    onSubmit={() => {
                      void doSave().catch((err) => toast.danger(err instanceof Error ? err.message : String(err)));
                    }}
                    showRefreshKey
                    forbidEnterSubmit
                  />
                ) : null}
              </Modal.Body>
              <Modal.Footer>
                <Button variant="secondary" slot="close">
                  取消
                </Button>
                <Button type="submit" form="form-edit" variant="primary">
                  保存
                </Button>
              </Modal.Footer>
            </Modal.Dialog>
          </Modal.Container>
        </Modal.Backdrop>
      </Modal>
    </>
  );
}
