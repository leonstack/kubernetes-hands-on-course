// edge-scheduler.go
package scheduler

import (
    "context"
    "fmt"
    "math"
    "strconv"
    "time"
    
    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/klog/v2"
)

const (
    EdgeZoneLabel        = "node.kubernetes.io/edge-zone"
    EdgeLatencyLabel     = "node.kubernetes.io/edge-latency"
    EdgeBandwidthLabel   = "node.kubernetes.io/edge-bandwidth"
    EdgeReliabilityLabel = "node.kubernetes.io/edge-reliability"
    
    PodEdgeZoneAnnotation     = "scheduler.kubernetes.io/edge-zone"
    PodLatencyRequirement     = "scheduler.kubernetes.io/max-latency"
    PodBandwidthRequirement   = "scheduler.kubernetes.io/min-bandwidth"
    PodReliabilityRequirement = "scheduler.kubernetes.io/min-reliability"
)

type EdgeScheduler struct {
    client      kubernetes.Interface
    edgeZones   map[string]*EdgeZone
    nodeMetrics map[string]*EdgeNodeMetrics
}

type EdgeZone struct {
    Name         string
    Location     GeoLocation
    Nodes        []string
    Connectivity ConnectivityInfo
    Resources    EdgeResourceInfo
}

type GeoLocation struct {
    Latitude  float64
    Longitude float64
    Region    string
    Country   string
}

type ConnectivityInfo struct {
    Latency     time.Duration
    Bandwidth   int64 // Mbps
    Reliability float64 // 0-1
    Jitter      time.Duration
}

type EdgeResourceInfo struct {
    TotalCPU    int64
    TotalMemory int64
    TotalStorage int64
    UsedCPU     int64
    UsedMemory  int64
    UsedStorage int64
}

type EdgeNodeMetrics struct {
    NodeName        string
    EdgeZone        string
    Latency         time.Duration
    Bandwidth       int64
    Reliability     float64
    ResourceUsage   EdgeResourceUsage
    NetworkQuality  NetworkQuality
    LastUpdated     time.Time
}

type EdgeResourceUsage struct {
    CPUUsage    float64
    MemoryUsage float64
    StorageUsage float64
    NetworkUsage float64
}

type NetworkQuality struct {
    PacketLoss   float64
    Jitter       time.Duration
    Throughput   int64
    RTT          time.Duration
}

func NewEdgeScheduler(client kubernetes.Interface) *EdgeScheduler {
    es := &EdgeScheduler{
        client:      client,
        edgeZones:   make(map[string]*EdgeZone),
        nodeMetrics: make(map[string]*EdgeNodeMetrics),
    }
    
    es.initializeEdgeZones()
    return es
}

func (es *EdgeScheduler) initializeEdgeZones() {
    // 初始化边缘区域配置
    es.edgeZones["edge-zone-east"] = &EdgeZone{
        Name: "edge-zone-east",
        Location: GeoLocation{
            Latitude:  40.7128,
            Longitude: -74.0060,
            Region:    "us-east",
            Country:   "US",
        },
        Connectivity: ConnectivityInfo{
            Latency:     10 * time.Millisecond,
            Bandwidth:   1000, // 1Gbps
            Reliability: 0.99,
            Jitter:      2 * time.Millisecond,
        },
    }
    
    es.edgeZones["edge-zone-west"] = &EdgeZone{
        Name: "edge-zone-west",
        Location: GeoLocation{
            Latitude:  37.7749,
            Longitude: -122.4194,
            Region:    "us-west",
            Country:   "US",
        },
        Connectivity: ConnectivityInfo{
            Latency:     15 * time.Millisecond,
            Bandwidth:   500, // 500Mbps
            Reliability: 0.95,
            Jitter:      5 * time.Millisecond,
        },
    }
    
    es.edgeZones["edge-zone-europe"] = &EdgeZone{
        Name: "edge-zone-europe",
        Location: GeoLocation{
            Latitude:  51.5074,
            Longitude: -0.1278,
            Region:    "eu-west",
            Country:   "UK",
        },
        Connectivity: ConnectivityInfo{
            Latency:     25 * time.Millisecond,
            Bandwidth:   300, // 300Mbps
            Reliability: 0.92,
            Jitter:      8 * time.Millisecond,
        },
    }
}

func (es *EdgeScheduler) SchedulePod(ctx context.Context, pod *v1.Pod) (string, error) {
    // 获取Pod的边缘调度要求
    requirements := es.extractPodRequirements(pod)
    
    // 获取候选节点
    candidateNodes, err := es.getCandidateNodes(ctx, requirements)
    if err != nil {
        return "", err
    }
    
    if len(candidateNodes) == 0 {
        return "", fmt.Errorf("no suitable edge nodes found for pod %s/%s", pod.Namespace, pod.Name)
    }
    
    // 计算节点分数
    nodeScores := es.scoreNodes(candidateNodes, requirements)
    
    // 选择最佳节点
    bestNode := es.selectBestNode(nodeScores)
    
    klog.Infof("Selected edge node %s for pod %s/%s", bestNode, pod.Namespace, pod.Name)
    return bestNode, nil
}

type PodEdgeRequirements struct {
    PreferredZone    string
    MaxLatency       time.Duration
    MinBandwidth     int64
    MinReliability   float64
    ResourceRequests v1.ResourceList
    LocationHint     *GeoLocation
}

func (es *EdgeScheduler) extractPodRequirements(pod *v1.Pod) *PodEdgeRequirements {
    req := &PodEdgeRequirements{
        ResourceRequests: make(v1.ResourceList),
    }
    
    // 提取注解中的要求
    if zone, exists := pod.Annotations[PodEdgeZoneAnnotation]; exists {
        req.PreferredZone = zone
    }
    
    if latencyStr, exists := pod.Annotations[PodLatencyRequirement]; exists {
        if latency, err := time.ParseDuration(latencyStr); err == nil {
            req.MaxLatency = latency
        }
    }
    
    if bandwidthStr, exists := pod.Annotations[PodBandwidthRequirement]; exists {
        if bandwidth, err := strconv.ParseInt(bandwidthStr, 10, 64); err == nil {
            req.MinBandwidth = bandwidth
        }
    }
    
    if reliabilityStr, exists := pod.Annotations[PodReliabilityRequirement]; exists {
        if reliability, err := strconv.ParseFloat(reliabilityStr, 64); err == nil {
            req.MinReliability = reliability
        }
    }
    
    // 聚合资源请求
    for _, container := range pod.Spec.Containers {
        for resource, quantity := range container.Resources.Requests {
            if existing, exists := req.ResourceRequests[resource]; exists {
                existing.Add(quantity)
                req.ResourceRequests[resource] = existing
            } else {
                req.ResourceRequests[resource] = quantity
            }
        }
    }
    
    return req
}

func (es *EdgeScheduler) getCandidateNodes(ctx context.Context, req *PodEdgeRequirements) ([]*v1.Node, error) {
    nodeList, err := es.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{
        LabelSelector: EdgeZoneLabel,
    })
    if err != nil {
        return nil, err
    }
    
    var candidates []*v1.Node
    
    for i := range nodeList.Items {
        node := &nodeList.Items[i]
        
        // 检查节点是否满足基本要求
        if es.nodeMatchesRequirements(node, req) {
            candidates = append(candidates, node)
        }
    }
    
    return candidates, nil
}

func (es *EdgeScheduler) nodeMatchesRequirements(node *v1.Node, req *PodEdgeRequirements) bool {
    // 检查边缘区域
    if req.PreferredZone != "" {
        if nodeZone, exists := node.Labels[EdgeZoneLabel]; !exists || nodeZone != req.PreferredZone {
            return false
        }
    }
    
    // 检查延迟要求
    if req.MaxLatency > 0 {
        if latencyStr, exists := node.Labels[EdgeLatencyLabel]; exists {
            if latency, err := time.ParseDuration(latencyStr); err == nil {
                if latency > req.MaxLatency {
                    return false
                }
            }
        }
    }
    
    // 检查带宽要求
    if req.MinBandwidth > 0 {
        if bandwidthStr, exists := node.Labels[EdgeBandwidthLabel]; exists {
            if bandwidth, err := strconv.ParseInt(bandwidthStr, 10, 64); err == nil {
                if bandwidth < req.MinBandwidth {
                    return false
                }
            }
        }
    }
    
    // 检查可靠性要求
    if req.MinReliability > 0 {
        if reliabilityStr, exists := node.Labels[EdgeReliabilityLabel]; exists {
            if reliability, err := strconv.ParseFloat(reliabilityStr, 64); err == nil {
                if reliability < req.MinReliability {
                    return false
                }
            }
        }
    }
    
    // 检查资源可用性
    return es.nodeHasEnoughResources(node, req.ResourceRequests)
}

func (es *EdgeScheduler) nodeHasEnoughResources(node *v1.Node, requests v1.ResourceList) bool {
    for resource, requested := range requests {
        if available, exists := node.Status.Allocatable[resource]; exists {
            if requested.Cmp(available) > 0 {
                return false
            }
        } else {
            return false
        }
    }
    return true
}

type EdgeNodeScore struct {
    NodeName string
    Score    float64
    Details  ScoreDetails
}

type ScoreDetails struct {
    LatencyScore     float64
    BandwidthScore   float64
    ReliabilityScore float64
    ResourceScore    float64
    LocationScore    float64
}

func (es *EdgeScheduler) scoreNodes(nodes []*v1.Node, req *PodEdgeRequirements) []EdgeNodeScore {
    var scores []EdgeNodeScore
    
    for _, node := range nodes {
        score := es.calculateNodeScore(node, req)
        scores = append(scores, score)
    }
    
    return scores
}

func (es *EdgeScheduler) calculateNodeScore(node *v1.Node, req *PodEdgeRequirements) EdgeNodeScore {
    details := ScoreDetails{}
    
    // 延迟分数 (权重: 30%)
    details.LatencyScore = es.calculateLatencyScore(node, req.MaxLatency)
    
    // 带宽分数 (权重: 25%)
    details.BandwidthScore = es.calculateBandwidthScore(node, req.MinBandwidth)
    
    // 可靠性分数 (权重: 20%)
    details.ReliabilityScore = es.calculateReliabilityScore(node, req.MinReliability)
    
    // 资源分数 (权重: 15%)
    details.ResourceScore = es.calculateResourceScore(node, req.ResourceRequests)
    
    // 位置分数 (权重: 10%)
    details.LocationScore = es.calculateLocationScore(node, req.LocationHint)
    
    // 加权总分
    totalScore := details.LatencyScore*0.3 +
                  details.BandwidthScore*0.25 +
                  details.ReliabilityScore*0.2 +
                  details.ResourceScore*0.15 +
                  details.LocationScore*0.1
    
    return EdgeNodeScore{
        NodeName: node.Name,
        Score:    totalScore,
        Details:  details,
    }
}

func (es *EdgeScheduler) calculateLatencyScore(node *v1.Node, maxLatency time.Duration) float64 {
    if maxLatency == 0 {
        return 100.0 // 没有延迟要求
    }
    
    latencyStr, exists := node.Labels[EdgeLatencyLabel]
    if !exists {
        return 50.0 // 默认分数
    }
    
    latency, err := time.ParseDuration(latencyStr)
    if err != nil {
        return 50.0
    }
    
    // 延迟越低分数越高
    if latency <= maxLatency {
        ratio := float64(latency) / float64(maxLatency)
        return 100.0 * (1.0 - ratio)
    }
    
    return 0.0 // 超过最大延迟要求
}

func (es *EdgeScheduler) calculateBandwidthScore(node *v1.Node, minBandwidth int64) float64 {
    if minBandwidth == 0 {
        return 100.0
    }
    
    bandwidthStr, exists := node.Labels[EdgeBandwidthLabel]
    if !exists {
        return 50.0
    }
    
    bandwidth, err := strconv.ParseInt(bandwidthStr, 10, 64)
    if err != nil {
        return 50.0
    }
    
    if bandwidth >= minBandwidth {
        // 带宽越高分数越高，但有上限
        ratio := float64(bandwidth) / float64(minBandwidth)
        return math.Min(100.0, 50.0+50.0*math.Log10(ratio))
    }
    
    return 0.0
}

func (es *EdgeScheduler) calculateReliabilityScore(node *v1.Node, minReliability float64) float64 {
    if minReliability == 0 {
        return 100.0
    }
    
    reliabilityStr, exists := node.Labels[EdgeReliabilityLabel]
    if !exists {
        return 50.0
    }
    
    reliability, err := strconv.ParseFloat(reliabilityStr, 64)
    if err != nil {
        return 50.0
    }
    
    if reliability >= minReliability {
        return 100.0 * reliability
    }
    
    return 0.0
}

func (es *EdgeScheduler) calculateResourceScore(node *v1.Node, requests v1.ResourceList) float64 {
    if len(requests) == 0 {
        return 100.0
    }
    
    var totalScore float64
    var resourceCount int
    
    for resource, requested := range requests {
        if allocatable, exists := node.Status.Allocatable[resource]; exists {
            ratio := float64(requested.MilliValue()) / float64(allocatable.MilliValue())
            // 资源利用率越低分数越高
            resourceScore := 100.0 * (1.0 - ratio)
            totalScore += math.Max(0, resourceScore)
            resourceCount++
        }
    }
    
    if resourceCount == 0 {
        return 50.0
    }
    
    return totalScore / float64(resourceCount)
}

func (es *EdgeScheduler) calculateLocationScore(node *v1.Node, locationHint *GeoLocation) float64 {
    if locationHint == nil {
        return 100.0
    }
    
    // 基于边缘区域计算位置分数
    zoneLabel, exists := node.Labels[EdgeZoneLabel]
    if !exists {
        return 50.0
    }
    
    zone, exists := es.edgeZones[zoneLabel]
    if !exists {
        return 50.0
    }
    
    // 计算地理距离
    distance := es.calculateDistance(locationHint, &zone.Location)
    
    // 距离越近分数越高
    maxDistance := 10000.0 // 10000km
    if distance <= maxDistance {
        return 100.0 * (1.0 - distance/maxDistance)
    }
    
    return 0.0
}

func (es *EdgeScheduler) calculateDistance(loc1, loc2 *GeoLocation) float64 {
    // 使用Haversine公式计算地理距离
    const earthRadius = 6371 // km
    
    lat1Rad := loc1.Latitude * math.Pi / 180
    lat2Rad := loc2.Latitude * math.Pi / 180
    deltaLatRad := (loc2.Latitude - loc1.Latitude) * math.Pi / 180
    deltaLonRad := (loc2.Longitude - loc1.Longitude) * math.Pi / 180
    
    a := math.Sin(deltaLatRad/2)*math.Sin(deltaLatRad/2) +
        math.Cos(lat1Rad)*math.Cos(lat2Rad)*
        math.Sin(deltaLonRad/2)*math.Sin(deltaLonRad/2)
    
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    
    return earthRadius * c
}

func (es *EdgeScheduler) selectBestNode(scores []EdgeNodeScore) string {
    if len(scores) == 0 {
        return ""
    }
    
    bestScore := scores[0]
    for _, score := range scores[1:] {
        if score.Score > bestScore.Score {
            bestScore = score
        }
    }
    
    return bestScore.NodeName
}

func (es *EdgeScheduler) UpdateNodeMetrics(nodeName string, metrics *EdgeNodeMetrics) {
    metrics.LastUpdated = time.Now()
    es.nodeMetrics[nodeName] = metrics
}

func (es *EdgeScheduler) GetEdgeZoneStatus(zoneName string) (*EdgeZone, error) {
    zone, exists := es.edgeZones[zoneName]
    if !exists {
        return nil, fmt.Errorf("edge zone %s not found", zoneName)
    }
    
    return zone, nil
}