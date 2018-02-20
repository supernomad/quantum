// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

node {
    try {
        go_dir = "/opt/go/src/github.com/supernomad/quantum"

        stage("Checkout") {
            checkout([
                $class: 'GitSCM',
                branches: scm.branches,
                doGenerateSubmoduleConfigurations: false,
                extensions: scm.extensions + [[$class: 'SubmoduleOption', parentCredentials: true, recursiveSubmodules: true]],
                userRemoteConfigs: scm.userRemoteConfigs
            ])
        }

        builder = docker.build('builder', '--network host --pull -f ./dist/docker/Dockerfile.builder ./dist/')
        builder.inside('--net host --cap-add NET_ADMIN --cap-add NET_RAW') {
            stage("Setup") {
                sh """
                    mkdir -p /opt/go/src/github.com/supernomad
                    ln -s ${env.WORKSPACE} ${go_dir}
                """
                sh "cd ${go_dir}; make setup_ci"
            }

            stage("Compile") {
                sh "cd ${go_dir}; make"
            }

            stage("Test") {
                sh "cd ${go_dir}; make CI=true check"
            }

            stage('Results') {
                junit allowEmptyResults: true, healthScaleFactor: 100.0, testResults: 'build_output/tests.xml'
                plot csvFileName: 'plot-d4a3949c-8a02-4ae6-9a18-ceb4f1e10b09.csv', exclZero: false, group: 'benchmarks', keepRecords: false, logarithmic: false, numBuilds: '10', style: 'line', title: 'Allocations Per Call', useDescr: false, xmlSeries: [[file: 'build_output/benchmarks.xml', nodeType: 'NODESET', url: '', xpath: '/Benchmarks/AllocsPerOp/*']], yaxis: 'Allocations (count)', yaxisMaximum: '', yaxisMinimum: ''
                plot csvFileName: 'plot-d4a3949c-8a02-4ae6-9a18-ceb4f1e10b10.csv', exclZero: false, group: 'benchmarks', keepRecords: false, logarithmic: false, numBuilds: '10', style: 'line', title: 'Allocated Bytes Per Call', useDescr: false, xmlSeries: [[file: 'build_output/benchmarks.xml', nodeType: 'NODESET', url: '', xpath: '/Benchmarks/AllocsBytesPerOp/*']], yaxis: 'Allocations (B)', yaxisMaximum: '', yaxisMinimum: ''
                plot csvFileName: 'plot-d4a3949c-8a02-4ae6-9a18-ceb4f1e10b11.csv', exclZero: false, group: 'benchmarks', keepRecords: false, logarithmic: false, numBuilds: '10', style: 'line', title: 'Time Per Call', useDescr: false, xmlSeries: [[file: 'build_output/benchmarks.xml', nodeType: 'NODESET', url: '', xpath: '/Benchmarks/NsPerOp/*']], yaxis: 'Time (ns)', yaxisMaximum: '', yaxisMinimum: ''
                cobertura autoUpdateHealth: false, autoUpdateStability: false, coberturaReportFile: 'build_output/coverage.xml', conditionalCoverageTargets: '70, 0, 0', failUnhealthy: false, failUnstable: false, lineCoverageTargets: '80, 0, 0', maxNumberOfBuilds: 0, methodCoverageTargets: '80, 0, 0', onlyStable: false, sourceEncoding: 'ASCII', zoomCoverageChart: false
                archiveArtifacts 'build_output/*.xml, quantum'
            }
        }
    }
    finally {
        cleanWs()
    }
}
