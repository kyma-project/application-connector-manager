package controllers

//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=applicationconnectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=applicationconnectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=applicationconnectors/finalizers,verbs=update

// Application Connector charts
//+kubebuilder:rbac:groups="applicationconnector.kyma-project.io",resources=applications,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups="compass.kyma-project.io",resources=compassconnections,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=create;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=list;get;patch
//+kubebuilder:rbac:groups="",resources=nodes,verbs=list;get
//+kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=list;get
//+kubebuilder:rbac:groups="metrics.k8s.io",resources=nodes,verbs=list;get
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;delete;update;patch
//+kubebuilder:rbac:groups="",resources=secrets,resourceNames=compass-agent-configuration;cluster-client-certificates,verbs=get;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;delete;update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;delete;update;patch
//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=list;watch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings,verbs=list;get;create;update;patch;delete;watch
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles;rolebindings,verbs=list;get;create;update;patch;delete;watch
//+kubebuilder:rbac:groups="",resources=limitranges,verbs=list;get;create;update;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=list;get;create;update;patch;delete;watch
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=list;get;create;update;patch;delete;watch
//+kubebuilder:rbac:groups=networking.istio.io,resources=gateways,verbs=list;get;create;update;patch;delete;watch
//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=list;get;create;update;patch;delete;watch
//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=list;watch;delete;create;patch;get
//+kubebuilder:rbac:groups=scheduling.k8s.io,resources=priorityclasses,verbs=list;get;create;update;patch;delete;watch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;delete;update;patch
