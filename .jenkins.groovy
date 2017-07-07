node {
    try {
        go_dir = "/opt/go/src/github.com/Supernomad/quantum"

        stage("Checkout") {
            checkout([
                $class: 'GitSCM',
                branches: scm.branches,
                doGenerateSubmoduleConfigurations: false,
                extensions: scm.extensions + [[$class: 'SubmoduleOption', parentCredentials: true, recursiveSubmodules: true]],
                userRemoteConfigs: scm.userRemoteConfigs
            ])
        }

        builder = docker.build('builder', '--pull -f ./dist/Dockerfile.builder ./dist/')
        builder.inside('--net host') {
            stage("Configure") {
                sh """
                    mkdir -p /opt/go/src/github.com/Supernomad
                    ln -s ${env.WORKSPACE} ${go_dir}
                """
            }

            stage("Dependencies") {
                sh "cd ${go_dir}; make setup_ci"
            }

            stage("Lint") {
                sh "cd ${go_dir}; make lint"
            }

            stage("Compile") {
                sh "cd ${go_dir}; make compile"
            }

            stage("Test") {
                sh "cd ${go_dir}; make test_ci"
            }

            stage('Results') {
                junit allowEmptyResults: true, testResults: 'tests.xml'
                step([$class: 'PlotBuilder', csvFileName: 'plot-56564010.csv', exclZero: false, group: 'benchmarks', keepRecords: false, logarithmic: false, numBuilds: '', style: 'line', title: 'Benchmarks', useDescr: false, xmlSeries: [[file: 'benchmarks.xml', nodeType: 'NODESET', url: '', xpath: '/Benchmarks/AllocsPerOp/*']], yaxis: 'Allocs', yaxisMaximum: '', yaxisMinimum: ''])
                step([$class: 'CoberturaPublisher', autoUpdateHealth: false, autoUpdateStability: false, coberturaReportFile: 'coverage.xml', failUnhealthy: false, failUnstable: false, maxNumberOfBuilds: 0, onlyStable: false, sourceEncoding: 'ASCII', zoomCoverageChart: false])
                archiveArtifacts '*.xml, quantum'
            }
        }
    }
    finally {
        stage("Cleanup") {
            cleanWs()
        }
    }
}
