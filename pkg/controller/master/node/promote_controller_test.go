package node

import (
	"reflect"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeBuilder struct {
	node *corev1.Node
}

func NewDefaultNodeBuilder() *NodeBuilder {
	return &NodeBuilder{
		node: &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "",
				Labels:            map[string]string{},
				Annotations:       map[string]string{},
				CreationTimestamp: metav1.NewTime(time.Now()),
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{},
			},
		},
	}
}

func (n *NodeBuilder) Name(name string) *NodeBuilder {
	n.node.Name = name
	return n
}

func (n *NodeBuilder) Harvester() *NodeBuilder {
	n.node.Labels[HarvesterManagedNodeLabelKey] = "true"
	return n
}

func (n *NodeBuilder) Management() *corev1.Node {
	n.node.Labels[KubeMasterNodeLabelKey] = "true"
	n.node.CreationTimestamp = metav1.NewTime(time.Now())
	return n.node
}

func (n *NodeBuilder) Worker() *corev1.Node {
	n.node.CreationTimestamp = metav1.NewTime(time.Now())
	return n.node
}

func (n *NodeBuilder) Running() *NodeBuilder {
	n.node.Annotations[HarvesterPromoteStatusAnnotationKey] = PromoteStatusRunning
	return n
}

func (n *NodeBuilder) Complete() *NodeBuilder {
	n.node.Annotations[HarvesterPromoteStatusAnnotationKey] = PromoteStatusComplete
	return n
}

func (n *NodeBuilder) Failed() *NodeBuilder {
	n.node.Annotations[HarvesterPromoteStatusAnnotationKey] = PromoteStatusFailed
	return n
}

func (n *NodeBuilder) NotReady() *NodeBuilder {
	ready := corev1.NodeCondition{
		Type:   corev1.NodeReady,
		Status: corev1.ConditionFalse,
	}
	n.node.Status.Conditions = append(n.node.Status.Conditions, ready)
	return n
}

var (
	mu1 = NewDefaultNodeBuilder().Name("m-unmanaged-1").Management()

	m1 = NewDefaultNodeBuilder().Name("m-1").Harvester().Management()
	m2 = NewDefaultNodeBuilder().Name("m-2").Harvester().Management()
	m3 = NewDefaultNodeBuilder().Name("m-3").Harvester().Management()

	mc1 = NewDefaultNodeBuilder().Name("m-complete-1").Harvester().Complete().Management()

	wr1 = NewDefaultNodeBuilder().Name("w-running-1").Harvester().Running().Worker()
	wf1 = NewDefaultNodeBuilder().Name("w-failed-1").Harvester().Failed().Worker()

	wnr1 = NewDefaultNodeBuilder().Name("w-notready-1").Harvester().NotReady().Worker()
	wnr2 = NewDefaultNodeBuilder().Name("w-notready-2").Harvester().NotReady().Worker()

	w1 = NewDefaultNodeBuilder().Name("w-1").Harvester().Worker()
	w2 = NewDefaultNodeBuilder().Name("w-2").Harvester().Worker()
	w3 = NewDefaultNodeBuilder().Name("w-3").Harvester().Worker()
)

func Test_selectPromoteNode(t *testing.T) {
	type args struct {
		nodeList []*corev1.Node
	}
	tests := []struct {
		name string
		args args
		want *corev1.Node
	}{
		{
			name: "one management",
			args: args{
				nodeList: []*corev1.Node{m1},
			},
			want: nil,
		},
		{
			name: "one management and one worker",
			args: args{
				nodeList: []*corev1.Node{m1, w1},
			},
			want: nil,
		},
		{
			name: "one management and two worker",
			args: args{
				nodeList: []*corev1.Node{m1, w1, w2},
			},
			want: w1,
		},
		{
			name: "one management and three worker",
			args: args{
				nodeList: []*corev1.Node{m1, w1, w2, w3},
			},
			want: w1,
		},
		{
			name: "one management and one not ready worker",
			args: args{
				nodeList: []*corev1.Node{m1, wnr1},
			},
			want: nil,
		},
		{
			name: "one management and one not ready worker and one ready worker",
			args: args{
				nodeList: []*corev1.Node{m1, wnr1, w1},
			},
			want: nil,
		},
		{
			name: "one management and one not ready worker and two ready worker",
			args: args{
				nodeList: []*corev1.Node{m1, wnr1, w1, w2},
			},
			want: w1,
		},
		{
			name: "one management and two not ready worker and one ready worker",
			args: args{
				nodeList: []*corev1.Node{m1, wnr1, wnr2, w1},
			},
			want: nil,
		},
		{
			name: "one management and two not ready worker and two ready worker",
			args: args{
				nodeList: []*corev1.Node{m1, wnr1, wnr2, w1, w2},
			},
			want: w1,
		},
		{
			name: "two management",
			args: args{
				nodeList: []*corev1.Node{m1, m2},
			},
			want: nil,
		},
		{
			name: "two management and one worker",
			args: args{
				nodeList: []*corev1.Node{m1, m2, w1},
			},
			want: w1,
		},
		{
			name: "two management and two worker",
			args: args{
				nodeList: []*corev1.Node{m1, m2, w1, w2},
			},
			want: w1,
		},
		{
			name: "two management and three worker",
			args: args{
				nodeList: []*corev1.Node{m1, m2, w1, w2, w3},
			},
			want: w1,
		},
		{
			name: "one management and one promoting worker",
			args: args{
				nodeList: []*corev1.Node{m1, wr1},
			},
			want: nil,
		},
		{
			name: "one management and one promoting worker and one worker",
			args: args{
				nodeList: []*corev1.Node{m1, wr1, w1},
			},
			want: nil,
		},
		{
			name: "one management and one promoting worker and two worker",
			args: args{
				nodeList: []*corev1.Node{m1, wr1, w1, w2},
			},
			want: nil,
		},
		{
			name: "one management and one promoting worker and three worker",
			args: args{
				nodeList: []*corev1.Node{m1, wr1, w1, w2, w3},
			},
			want: nil,
		},
		{
			name: "one management and one promote failed worker",
			args: args{
				nodeList: []*corev1.Node{m1, wf1},
			},
			want: nil,
		},
		{
			name: "one management and one promote failed worker and one worker",
			args: args{
				nodeList: []*corev1.Node{m1, wf1, w1},
			},
			want: nil,
		},
		{
			name: "one management and one promote failed worker and two worker",
			args: args{
				nodeList: []*corev1.Node{m1, wf1, w1, w2},
			},
			want: nil,
		},
		{
			name: "one management and one promote failed worker and three worker",
			args: args{
				nodeList: []*corev1.Node{m1, wf1, w1, w2, w3},
			},
			want: nil,
		},
		{
			name: "one management and one promoted management",
			args: args{
				nodeList: []*corev1.Node{m1, mc1},
			},
			want: nil,
		},
		{
			name: "one management and one promoted management and one worker",
			args: args{
				nodeList: []*corev1.Node{m1, mc1, w1},
			},
			want: w1,
		},
		{
			name: "one management and one promoted management and two worker",
			args: args{
				nodeList: []*corev1.Node{m1, mc1, w1, w2},
			},
			want: w1,
		},
		{
			name: "one management and one promoted management and three worker",
			args: args{
				nodeList: []*corev1.Node{m1, mc1, w1, w2, w3},
			},
			want: w1,
		},
		{
			name: "one managed management and one unmanaged management",
			args: args{
				nodeList: []*corev1.Node{m1, mu1},
			},
			want: nil,
		},
		{
			name: "one managed management and one unmanaged management and one worker",
			args: args{
				nodeList: []*corev1.Node{m1, mu1, w1},
			},
			want: w1,
		},
		{
			name: "one managed management and one unmanaged management and two worker",
			args: args{
				nodeList: []*corev1.Node{m1, mu1, w1, w2},
			},
			want: w1,
		},
		{
			name: "one managed management and one unmanaged management and three worker",
			args: args{
				nodeList: []*corev1.Node{m1, mu1, w1, w2, w3},
			},
			want: w1,
		},
		{
			name: "three management",
			args: args{
				nodeList: []*corev1.Node{m1, m2, m3},
			},
			want: nil,
		},
		{
			name: "three management and one worker",
			args: args{
				nodeList: []*corev1.Node{m1, m2, m3, w1},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectPromoteNode(tt.args.nodeList); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("selectPromoteNode() = %v, want %v", got, tt.want)
			}
		})
	}
}