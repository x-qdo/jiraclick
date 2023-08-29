@Library('pipeline-common@master') _

def projectName = 'jiraclick'

properties(skyDeployer.wrapProperties())

withResultReporting(slackChannel: '#ci') {
  inDockerAgent('dockerBuildEcr') {
    checkout(scm)

    echo("start deploy")
    skyDeployer.deployOnCommentTrigger(
      kubernetesDeployment: projectName,
      kubernetesNamespace: 'slack-bot',
      lockGlobally: false,
      deployMap: [
        'beta': [
          'preCheck': false,
          'env': 'beta',
        ],
      ],
      checklistFor: { env ->
        [[
          name: 'OK?',
          description: "Are you feeling good about this change in ${env.name}?"
        ]]
      },
    )
  }
}
