node {
    stage ("Checkout") {
        checkout scm
        //checkout([
        //    $class: 'GitSCM',
        //    branches: scm.branches,
        //    doGenerateSubmoduleConfigurations: false,
        //    extensions: scm.extensions + [[$class: 'SubmoduleOption', parentCredentials: true, recursiveSubmodules: true]],
        //    userRemoteConfigs: scm.userRemoteConfigs
        //])
    }
}
